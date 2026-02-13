# BLE GATT 操作ガイド

`tinygo.org/x/bluetooth` ライブラリを使用した BLE (Bluetooth Low Energy) GATT 操作の手順書です。

## 目次

- [前提知識](#前提知識)
- [共通: アダプタの初期化](#共通-アダプタの初期化)
- [セントラル (クライアント) 操作](#セントラル-クライアント-操作)
  - [1. デバイスのスキャン](#1-デバイスのスキャン)
  - [2. デバイスへの接続](#2-デバイスへの接続)
  - [3. サービスの探索](#3-サービスの探索)
  - [4. キャラクタリスティックの探索](#4-キャラクタリスティックの探索)
  - [5. キャラクタリスティックの読み取り](#5-キャラクタリスティックの読み取り)
  - [6. キャラクタリスティックへの書き込み](#6-キャラクタリスティックへの書き込み)
  - [7. 通知 (Notification) の受信](#7-通知-notification-の受信)
  - [8. 切断](#8-切断)
- [ペリフェラル (サーバー) 操作](#ペリフェラル-サーバー-操作)
  - [1. アドバタイズの設定と開始](#1-アドバタイズの設定と開始)
  - [2. GATT サービスの定義と登録](#2-gatt-サービスの定義と登録)
  - [3. クライアントからの書き込みの処理](#3-クライアントからの書き込みの処理)
  - [4. 通知 (Notification) の送信](#4-通知-notification-の送信)
- [キャラクタリスティック権限一覧](#キャラクタリスティック権限一覧)
- [定義済み UUID](#定義済み-uuid)
- [セントラル操作の完全なサンプル](#セントラル操作の完全なサンプル)
- [ペリフェラル操作の完全なサンプル](#ペリフェラル操作の完全なサンプル)

---

## 前提知識

BLE の GATT (Generic Attribute Profile) は、デバイス間のデータ交換を定義するプロトコルです。

| 用語 | 説明 |
|------|------|
| **セントラル (Central)** | 他のデバイスに接続してデータを読み書きする側 (スマートフォン等) |
| **ペリフェラル (Peripheral)** | アドバタイズを行い、接続を待ち受ける側 (センサー等) |
| **サービス (Service)** | 関連するデータをグループ化する単位。UUID で識別される |
| **キャラクタリスティック (Characteristic)** | サービス内の個々のデータポイント。UUID で識別される |
| **通知 (Notification)** | ペリフェラルからセントラルへの非同期データ送信 |

---

## 共通: アダプタの初期化

すべての BLE 操作の前に、アダプタを有効化する必要があります。

```go
import "tinygo.org/x/bluetooth"

var adapter = bluetooth.DefaultAdapter

func main() {
    // BLE スタックを有効化
    err := adapter.Enable()
    if err != nil {
        println("BLE の有効化に失敗:", err.Error())
        return
    }
}
```

---

## セントラル (クライアント) 操作

セントラルは、ペリフェラルのスキャン → 接続 → サービス/キャラクタリスティック探索 → データ読み書き という流れで操作します。

### 1. デバイスのスキャン

`Adapter.Scan` を呼び出すと、周囲の BLE デバイスを検出するたびにコールバックが実行されます。`Scan` はブロッキング関数であり、`StopScan` が呼ばれるまで制御を返しません。

```go
// すべてのデバイスをスキャンして表示
err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
    println("発見:", result.Address.String(), result.RSSI, result.LocalName())
})
```

特定のデバイスを見つけたらスキャンを停止するには、チャネルとの組み合わせが有効です。

```go
ch := make(chan bluetooth.ScanResult, 1)

err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
    // MACアドレスで特定のデバイスを探す
    if result.Address.String() == "EE:74:7D:C9:2A:68" {
        adapter.StopScan()
        ch <- result
    }
})
```

アドバタイズに含まれるサービス UUID でフィルタリングすることもできます。

```go
err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
    // 特定のサービスをアドバタイズしているデバイスのみ対象にする
    if !result.AdvertisementPayload.HasServiceUUID(bluetooth.ServiceUUIDHeartRate) {
        return
    }
    adapter.StopScan()
    ch <- result
})
```

`ScanResult` から取得できる情報:

| フィールド/メソッド | 型 | 説明 |
|---|---|---|
| `Address` | `Address` | デバイスの MAC アドレス |
| `RSSI` | `int16` | 受信信号強度 (dBm) |
| `LocalName()` | `string` | デバイスのローカル名 |
| `AdvertisementPayload.HasServiceUUID(uuid)` | `bool` | 指定 UUID のサービスを含むか |
| `AdvertisementPayload.ServiceUUIDs()` | `[]UUID` | アドバタイズされているサービス UUID 一覧 |
| `AdvertisementPayload.ManufacturerData()` | `[]ManufacturerDataElement` | メーカー固有データ |

### 2. デバイスへの接続

スキャンで見つけたデバイスに `Adapter.Connect` で接続します。

```go
result := <-ch
device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
if err != nil {
    println("接続失敗:", err.Error())
    return
}
println("接続完了:", result.Address.String())
```

`ConnectionParams` には接続パラメータを指定できます。空の構造体を渡すとデフォルト値が使用されます。

```go
params := bluetooth.ConnectionParams{
    ConnectionTimeout: bluetooth.NewDuration(4 * time.Second),
    MinInterval:       bluetooth.NewDuration(7500 * time.Microsecond),
    MaxInterval:       bluetooth.NewDuration(15 * time.Millisecond),
    Timeout:           bluetooth.NewDuration(4 * time.Second),
}
device, err := adapter.Connect(result.Address, params)
```

### 3. サービスの探索

接続後、`Device.DiscoverServices` でペリフェラルが提供するサービスを探索します。

```go
// すべてのサービスを探索
services, err := device.DiscoverServices(nil)
if err != nil {
    println("サービス探索失敗:", err.Error())
    return
}

for _, service := range services {
    println("サービス:", service.UUID().String())
}
```

特定のサービスのみを探索する場合は、UUID を指定します。

```go
// Heart Rate サービスのみ探索
services, err := device.DiscoverServices([]bluetooth.UUID{
    bluetooth.ServiceUUIDHeartRate,
})
```

### 4. キャラクタリスティックの探索

`DeviceService.DiscoverCharacteristics` で、サービス内のキャラクタリスティックを探索します。

```go
// すべてのキャラクタリスティックを探索
chars, err := service.DiscoverCharacteristics(nil)
if err != nil {
    println("キャラクタリスティック探索失敗:", err.Error())
    return
}

for _, char := range chars {
    println("キャラクタリスティック:", char.UUID().String())
}
```

特定のキャラクタリスティックを探索する場合:

```go
chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{
    bluetooth.CharacteristicUUIDHeartRateMeasurement,
})
```

複数のキャラクタリスティックを同時に探索する場合、UUID の指定順に結果が返されます。

```go
chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{rxUUID, txUUID})
rx := chars[0] // rxUUID に対応
tx := chars[1] // txUUID に対応
```

### 5. キャラクタリスティックの読み取り

`DeviceCharacteristic.Read` で値を読み取ります。バッファを事前に確保して渡します。

```go
buf := make([]byte, 255)
n, err := char.Read(buf)
if err != nil {
    println("読み取り失敗:", err.Error())
    return
}
println("値:", string(buf[:n]))
```

MTU (Maximum Transmission Unit) の取得:

```go
mtu, err := char.GetMTU()
if err != nil {
    println("MTU 取得失敗:", err.Error())
} else {
    println("MTU:", mtu)
}
```

### 6. キャラクタリスティックへの書き込み

応答を待たない書き込み (`Write Without Response`):

```go
data := []byte("Hello BLE")
_, err := char.WriteWithoutResponse(data)
if err != nil {
    println("書き込み失敗:", err.Error())
}
```

MTU を超えるデータを送信する場合は、分割して送ります。

```go
sendbuf := []byte("long data to send...")
for len(sendbuf) > 0 {
    // 20 バイトずつ分割して送信
    partlen := 20
    if len(sendbuf) < partlen {
        partlen = len(sendbuf)
    }
    _, err := char.WriteWithoutResponse(sendbuf[:partlen])
    if err != nil {
        println("送信失敗:", err.Error())
        break
    }
    sendbuf = sendbuf[partlen:]
}
```

### 7. 通知 (Notification) の受信

`DeviceCharacteristic.EnableNotifications` でコールバックを登録すると、ペリフェラルからの通知を非同期に受信できます。

```go
err := char.EnableNotifications(func(buf []byte) {
    println("通知データ:", buf)
})
if err != nil {
    println("通知の有効化に失敗:", err.Error())
    return
}

// メインゴルーチンをブロックして通知を受け続ける
select {}
```

#### コールバック内での処理に関する注意

コールバックの実行コンテキストはプラットフォームによって異なります。ポータビリティと安全性のために、**コールバック内では最小限の処理に留め、チャネル経由で別の goroutine に処理を委譲する**ことを推奨します。

| プラットフォーム | コールバックの実行コンテキスト | 注意点 |
|---|---|---|
| macOS (Darwin) | 通知ごとに新しい goroutine を起動 | 比較的安全だが、goroutine が大量生成される可能性がある |
| Linux (BlueZ) | 専用リスナー goroutine 内で直接呼び出し | 処理中は次の通知がブロックされる |
| Windows | WinRT イベントハンドラ | システムスレッドプール上で実行される |
| HCI/NINA/CYW43439 | 専用ハンドラ goroutine 内で直接呼び出し | 処理中は次の通知がブロックされる |
| Nordic SoftDevice | **割り込みコンテキスト (ISR)** | **ヒープ割り当て不可、長時間の処理不可** |

特に Nordic SoftDevice では、コールバックが割り込みハンドラから直接呼ばれるため、`make` によるメモリ割り当てなど多くの操作が制限されます。

**推奨パターン: チャネルによる処理の委譲**

```go
ch := make(chan []byte, 1)

err := char.EnableNotifications(func(buf []byte) {
    // コールバック内ではデータのコピーと送信のみ行う
    // buf はコールバック終了後に再利用される可能性があるため必ずコピーする
    data := make([]byte, len(buf))
    copy(data, buf)
    select {
    case ch <- data:
    default:
        // チャネルが満杯なら破棄 (バッファサイズで調整可能)
    }
})
if err != nil {
    println("通知の有効化に失敗:", err.Error())
    return
}

// 別の goroutine で実際の処理を行う
go func() {
    for data := range ch {
        // ここで時間のかかる処理を安全に実行できる
        process(data)
    }
}()

// メインゴルーチンをブロック
select {}
```

> **Note:** `buf` の寿命はコールバックの呼び出し中のみ保証されます。コールバック外で使用する場合は必ず `copy` してください。

### 8. 切断

```go
err := device.Disconnect()
if err != nil {
    println("切断失敗:", err.Error())
}
```

---

## ペリフェラル (サーバー) 操作

ペリフェラルは、サービスの定義 → アドバタイズ開始 → クライアントからの接続待ち → データ提供 という流れです。

### 1. アドバタイズの設定と開始

ペリフェラルとして他のデバイスに発見されるために、アドバタイズを設定して開始します。

```go
adv := adapter.DefaultAdvertisement()
err := adv.Configure(bluetooth.AdvertisementOptions{
    LocalName:    "MyDevice",                                  // デバイス名
    ServiceUUIDs: []bluetooth.UUID{bluetooth.ServiceUUIDHeartRate}, // 提供サービス
})
if err != nil {
    println("アドバタイズ設定失敗:", err.Error())
    return
}

err = adv.Start()
if err != nil {
    println("アドバタイズ開始失敗:", err.Error())
    return
}
```

### 2. GATT サービスの定義と登録

`Adapter.AddService` でサービスとそのキャラクタリスティックを登録します。`Handle` フィールドに `Characteristic` のポインタを指定すると、登録後にそのキャラクタリスティックへの参照を保持できます。

```go
var myCharacteristic bluetooth.Characteristic

err := adapter.AddService(&bluetooth.Service{
    UUID: bluetooth.ServiceUUIDHeartRate,
    Characteristics: []bluetooth.CharacteristicConfig{
        {
            Handle: &myCharacteristic,                               // 後で値を更新するための参照
            UUID:   bluetooth.CharacteristicUUIDHeartRateMeasurement,
            Value:  []byte{0, 75},                                   // 初期値
            Flags:  bluetooth.CharacteristicNotifyPermission,        // 権限
        },
        {
            UUID:  bluetooth.CharacteristicUUIDBodySensorLocation,
            Value: []byte{1},                                        // "Chest"
            Flags: bluetooth.CharacteristicReadPermission,           // 読み取り専用
        },
    },
})
if err != nil {
    println("サービス登録失敗:", err.Error())
    return
}
```

### 3. クライアントからの書き込みの処理

`CharacteristicConfig.WriteEvent` にコールバックを設定すると、クライアントからの書き込みを処理できます。

```go
var rxChar bluetooth.Characteristic
var txChar bluetooth.Characteristic

err := adapter.AddService(&bluetooth.Service{
    UUID: bluetooth.ServiceUUIDNordicUART,
    Characteristics: []bluetooth.CharacteristicConfig{
        {
            Handle: &rxChar,
            UUID:   bluetooth.CharacteristicUUIDUARTRX,
            Flags:  bluetooth.CharacteristicWritePermission |
                    bluetooth.CharacteristicWriteWithoutResponsePermission,
            WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
                // クライアントからの書き込みデータを処理
                println("受信:", string(value))

                // エコーバック: 受け取ったデータを通知で返す
                txChar.Write(value)
            },
        },
        {
            Handle: &txChar,
            UUID:   bluetooth.CharacteristicUUIDUARTTX,
            Flags:  bluetooth.CharacteristicNotifyPermission |
                    bluetooth.CharacteristicReadPermission,
        },
    },
})
```

`WriteEvent` コールバックのパラメータ:

| パラメータ | 型 | 説明 |
|---|---|---|
| `client` | `Connection` | 書き込みを行ったクライアントの接続識別子 |
| `offset` | `int` | 書き込みオフセット (通常は 0) |
| `value` | `[]byte` | 書き込まれたデータ |

### 4. 通知 (Notification) の送信

`Characteristic.Write` でキャラクタリスティックの値を更新すると、通知を購読中のクライアントに自動的に通知が送信されます。

```go
// 値を更新して通知を送信
_, err := myCharacteristic.Write([]byte{0, 80}) // 心拍数 80bpm
if err != nil {
    println("通知送信失敗:", err.Error())
}
```

定期的に通知を送信する例:

```go
for {
    time.Sleep(time.Second)
    heartRate := readSensor() // センサーから値を取得
    myCharacteristic.Write([]byte{0, heartRate})
}
```

---

## キャラクタリスティック権限一覧

`CharacteristicConfig.Flags` に設定する権限フラグです。`|` 演算子で複数指定できます。

| 定数 | 説明 |
|------|------|
| `CharacteristicBroadcastPermission` | 値のブロードキャストを許可 |
| `CharacteristicReadPermission` | 値の読み取りを許可 |
| `CharacteristicWriteWithoutResponsePermission` | 応答なし書き込み (Write Command) を許可 |
| `CharacteristicWritePermission` | 応答あり書き込み (Write Request) を許可 |
| `CharacteristicNotifyPermission` | 通知 (Notification) を許可 |
| `CharacteristicIndicatePermission` | インジケーション (Indication) を許可 |

組み合わせ例:

```go
// 読み書き可能
Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission

// 書き込み (応答あり + 応答なし) 可能
Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission

// 通知と読み取りを許可
Flags: bluetooth.CharacteristicNotifyPermission | bluetooth.CharacteristicReadPermission
```

---

## 定義済み UUID

このライブラリには Bluetooth SIG 標準の UUID が定義済みです。

サービス UUID (`service_uuids.go`):

```go
bluetooth.ServiceUUIDHeartRate          // 0x180D - Heart Rate
bluetooth.ServiceUUIDBattery            // 0x180F - Battery Service
bluetooth.ServiceUUIDDeviceInformation  // 0x180A - Device Information
bluetooth.ServiceUUIDNordicUART         // Nordic UART Service (カスタム UUID)
```

キャラクタリスティック UUID (`characteristic_uuids.go`):

```go
bluetooth.CharacteristicUUIDHeartRateMeasurement  // 0x2A37
bluetooth.CharacteristicUUIDBodySensorLocation    // 0x2A38
bluetooth.CharacteristicUUIDBatteryLevel          // 0x2A19
bluetooth.CharacteristicUUIDUARTRX                // Nordic UART RX
bluetooth.CharacteristicUUIDUARTTX                // Nordic UART TX
```

カスタム UUID を作成する場合:

```go
// 16-bit UUID (Bluetooth SIG 標準)
uuid16 := bluetooth.New16BitUUID(0x180D)

// 128-bit UUID (カスタム)
uuid128 := bluetooth.NewUUID([16]byte{
    0x6E, 0x40, 0x00, 0x01, 0xB5, 0xA3, 0xF3, 0x93,
    0xE0, 0xA9, 0xE5, 0x0E, 0x24, 0xDC, 0xCA, 0x9E,
})
```

---

## セントラル操作の完全なサンプル

Heart Rate Monitor ペリフェラルに接続し、心拍数の通知を受信する例です。

```go
package main

import "tinygo.org/x/bluetooth"

var adapter = bluetooth.DefaultAdapter

func main() {
    must("enable BLE stack", adapter.Enable())

    // --- スキャン ---
    ch := make(chan bluetooth.ScanResult, 1)
    err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
        if result.AdvertisementPayload.HasServiceUUID(bluetooth.ServiceUUIDHeartRate) {
            adapter.StopScan()
            ch <- result
        }
    })
    must("start scan", err)

    // --- 接続 ---
    result := <-ch
    device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
    must("connect", err)
    println("接続完了:", result.Address.String())

    // --- サービス探索 ---
    services, err := device.DiscoverServices([]bluetooth.UUID{
        bluetooth.ServiceUUIDHeartRate,
    })
    must("discover services", err)

    // --- キャラクタリスティック探索 ---
    chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{
        bluetooth.CharacteristicUUIDHeartRateMeasurement,
    })
    must("discover characteristics", err)

    // --- 通知の受信 ---
    must("enable notifications", chars[0].EnableNotifications(func(buf []byte) {
        println("心拍数:", uint8(buf[1]), "bpm")
    }))

    // メインゴルーチンをブロック
    select {}
}

func must(action string, err error) {
    if err != nil {
        panic("failed to " + action + ": " + err.Error())
    }
}
```

---

## ペリフェラル操作の完全なサンプル

Heart Rate Service をアドバタイズし、定期的に通知を送信する例です。

```go
package main

import (
    "math/rand"
    "time"

    "tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
    must("enable BLE stack", adapter.Enable())

    // --- アドバタイズ ---
    adv := adapter.DefaultAdvertisement()
    must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
        LocalName:    "Go HRS",
        ServiceUUIDs: []bluetooth.UUID{bluetooth.ServiceUUIDHeartRate},
    }))
    must("start adv", adv.Start())

    // --- サービス登録 ---
    var heartRateMeasurement bluetooth.Characteristic

    must("add service", adapter.AddService(&bluetooth.Service{
        UUID: bluetooth.ServiceUUIDHeartRate,
        Characteristics: []bluetooth.CharacteristicConfig{
            {
                Handle: &heartRateMeasurement,
                UUID:   bluetooth.CharacteristicUUIDHeartRateMeasurement,
                Value:  []byte{0, 75},
                Flags:  bluetooth.CharacteristicNotifyPermission,
            },
            {
                UUID:  bluetooth.CharacteristicUUIDBodySensorLocation,
                Value: []byte{1}, // Chest
                Flags: bluetooth.CharacteristicReadPermission,
            },
        },
    }))

    // --- 定期的に通知を送信 ---
    for {
        time.Sleep(time.Second)
        heartRate := uint8(65 + rand.Intn(20))
        heartRateMeasurement.Write([]byte{0, heartRate})
        println("通知送信:", heartRate, "bpm")
    }
}

func must(action string, err error) {
    if err != nil {
        panic("failed to " + action + ": " + err.Error())
    }
}
```
