# クロスプラットフォームガイド (Windows / macOS / Linux)

`tinygo.org/x/bluetooth` ライブラリを使用して、Windows・macOS・Linux の 3 OS で共通に動作する BLE コードを書く際の注意点をまとめたガイドです。

## 目次

- [プラットフォーム別機能サポート表](#プラットフォーム別機能サポート表)
- [注意点](#注意点)
  - [1. macOS ではペリフェラル (サーバー) が動かない](#1-macos-ではペリフェラル-サーバー-が動かない)
  - [2. macOS のデバイスアドレスは MAC ではなく UUID](#2-macos-のデバイスアドレスは-mac-ではなく-uuid)
  - [3. Linux では Write (応答あり) が未実装](#3-linux-では-write-応答あり-が未実装)
  - [4. Adapter.Address() は Windows/macOS で未実装](#4-adapteraddress-は-windowsmacos-で未実装)
  - [5. RequestConnectionParams は全 OS で未実装](#5-requestconnectionparams-は全-os-で未実装)
  - [6. 通知モードの選択は Windows のみ](#6-通知モードの選択は-windows-のみ)
  - [7. スキャンの内部実装差異](#7-スキャンの内部実装差異)
- [3 OS 共通コードのチェックリスト](#3-os-共通コードのチェックリスト)
- [クロスプラットフォーム対応のサンプルコード](#クロスプラットフォーム対応のサンプルコード)

---

## プラットフォーム別機能サポート表

### セントラル (クライアント) 操作

| 機能 | Linux | macOS | Windows |
|------|:-----:|:-----:|:-------:|
| スキャン | o | o | o |
| 接続 | o | o | o |
| サービス探索 | o | o | o |
| キャラクタリスティック探索 | o | o | o |
| Read | o | o | o |
| WriteWithoutResponse | o | o | o |
| Write (応答あり) | **x** | o | o |
| 通知 (Notification) の受信 | o | o | o |
| EnableNotificationsWithMode | **x** | **x** | o |
| GetMTU | o | o | o |
| Disconnect | o | o | o |

### ペリフェラル (サーバー) 操作

| 機能 | Linux | macOS | Windows |
|------|:-----:|:-----:|:-------:|
| アドバタイズ | o | **x** | o |
| サービス登録 (AddService) | o | **x** | o |
| キャラクタリスティック登録 | o | **x** | o |
| 通知 (Notification) の送信 | o | **x** | o |
| WriteEvent コールバック | o | **x** | o |

### その他

| 機能 | Linux | macOS | Windows |
|------|:-----:|:-----:|:-------:|
| Adapter.Address() | o | **x** | **x** |
| RequestConnectionParams | **x** | **x** | **x** |
| デバイスアドレス形式 | MAC | UUID | MAC |

---

## 注意点

### 1. macOS ではペリフェラル (サーバー) が動かない

macOS の実装はセントラル (クライアント) 操作のみをサポートしています。以下の API は macOS で未実装です。

- `Adapter.AddService()`
- `Advertisement.Configure()` / `Advertisement.Start()`
- `Characteristic.Write()` (サーバー側の値更新)

**3 OS 共通で動作するのはセントラル操作のみです。** ペリフェラルとして動作させる必要がある場合は、Linux と Windows のみを対象にしてください。

### 2. macOS のデバイスアドレスは MAC ではなく UUID

macOS の CoreBluetooth はプライバシー保護のため、ペリフェラルの MAC アドレスを公開しません。代わりにランダムな UUID が割り当てられます。

```go
// Linux/Windows: "EE:74:7D:C9:2A:68" のような MAC アドレス
// macOS:         "A7F2D3B1-..." のような UUID 文字列
result.Address.String()
```

この UUID には以下の特性があります。

- セントラルごとに異なる UUID が同じデバイスに対して割り当てられる
- アプリ再起動で UUID が変わる可能性がある
- MAC アドレスとの相互変換はできない

**対策: MAC アドレスではなく `LocalName()` や サービス UUID でデバイスを識別する。**

```go
// 悪い例: MAC アドレス指定 (macOS で動かない)
if result.Address.String() == "EE:74:7D:C9:2A:68" {
    adapter.StopScan()
}

// 良い例: サービス UUID でフィルタ (全 OS で動く)
if result.AdvertisementPayload.HasServiceUUID(targetServiceUUID) {
    adapter.StopScan()
}

// 良い例: デバイス名でフィルタ (全 OS で動く)
if result.LocalName() == "MyDevice" {
    adapter.StopScan()
}
```

### 3. Linux では Write (応答あり) が未実装

Linux (BlueZ) の GATT クライアント実装には、応答を待つ `Write()` 相当のメソッドがありません。`WriteWithoutResponse` のみが利用可能です。

```go
// 全 OS で動く
char.WriteWithoutResponse(data)

// Linux で動かない (メソッドが存在しない)
// macOS / Windows でのみ動作する
```

3 OS 共通で書き込みを行う場合は `WriteWithoutResponse` を使用してください。ペリフェラル側のキャラクタリスティックにも対応する権限フラグを設定する必要があります。

```go
Flags: bluetooth.CharacteristicWriteWithoutResponsePermission
```

### 4. Adapter.Address() は Windows/macOS で未実装

自デバイスの Bluetooth アドレスを取得する `Adapter.Address()` は Linux でのみ動作します。

```go
addr, err := adapter.Address()
// Linux:   正常に MAC アドレスを返す
// Windows: "not implemented" エラー
// macOS:   "not implemented" エラー
```

自デバイスのアドレスに依存するロジックは避けてください。

### 5. RequestConnectionParams は全 OS で未実装

接続パラメータの変更要求は、すべてのデスクトッププラットフォームで未実装です。

```go
device.RequestConnectionParams(params) // 全 OS で何もしない (nil を返す)
```

`Adapter.Connect()` 呼び出し時の `ConnectionParams` は一部反映されますが、挙動は OS とドライバの実装に依存します。

### 6. 通知モードの選択は Windows のみ

Windows には `EnableNotificationsWithMode()` があり、Notification と Indication を明示的に選択できますが、macOS と Linux にはこのメソッドがありません。

3 OS 共通コードでは `EnableNotifications` のみを使用してください。

```go
// 全 OS で動く
char.EnableNotifications(func(buf []byte) {
    // 通知データの処理
})
```

### 7. スキャンの内部実装差異

各 OS のスキャン実装は内部的に異なる方式を採用しており、挙動に微妙な差が出ます。

| | Windows | macOS | Linux |
|---|---------|-------|-------|
| 内部方式 | WinRT イベント | CoreBluetooth デリゲート | D-Bus シグナル/ポーリング |
| 重複アドバタイズ | 受信する | 抑制される | ポーリングのため取りこぼしの可能性あり |
| バックエンド | WinRT BLE API | CoreBluetooth | BlueZ (D-Bus) |

Linux の D-Bus ベースのスキャンはプロパティ変更の監視方式のため、短時間しかアドバタイズしないデバイスを見逃す可能性があります。確実にデバイスを発見するため、スキャン時間に余裕を持たせてください。

---

## 3 OS 共通コードのチェックリスト

- [ ] セントラル操作のみを使っている (`AddService` / `Advertisement` を使わない)
- [ ] デバイスの識別に MAC アドレスを使わず、`LocalName()` か `HasServiceUUID()` でフィルタしている
- [ ] 書き込みには `WriteWithoutResponse` を使っている
- [ ] `Adapter.Address()` に依存していない
- [ ] 通知の受信には `EnableNotifications` を使っている (`EnableNotificationsWithMode` ではなく)
- [ ] `RequestConnectionParams` の成功に依存していない

---

## クロスプラットフォーム対応のサンプルコード

サービス UUID でデバイスを検索し、接続してデータを読み取り、通知を受信する例です。全ての OS で共通に動作します。

```go
package main

import "tinygo.org/x/bluetooth"

var adapter = bluetooth.DefaultAdapter

// 接続先のサービスとキャラクタリスティック UUID
var (
    targetServiceUUID = bluetooth.ServiceUUIDHeartRate
    targetCharUUID    = bluetooth.CharacteristicUUIDHeartRateMeasurement
)

func main() {
    // 1. BLE スタックを有効化
    must("enable BLE stack", adapter.Enable())

    // 2. サービス UUID でデバイスをスキャン
    //    (MAC アドレスではなくサービス UUID を使うことで macOS でも動作する)
    ch := make(chan bluetooth.ScanResult, 1)
    err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
        if result.AdvertisementPayload.HasServiceUUID(targetServiceUUID) {
            adapter.StopScan()
            ch <- result
        }
    })
    must("start scan", err)

    // 3. デバイスに接続
    //    (ConnectionParams は空のデフォルト値を使用)
    result := <-ch
    println("発見:", result.LocalName(), result.Address.String())
    device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
    must("connect", err)
    println("接続完了")

    // 4. サービスを探索
    services, err := device.DiscoverServices([]bluetooth.UUID{targetServiceUUID})
    must("discover services", err)
    if len(services) == 0 {
        panic("サービスが見つかりません")
    }

    // 5. キャラクタリスティックを探索
    chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{targetCharUUID})
    must("discover characteristics", err)
    if len(chars) == 0 {
        panic("キャラクタリスティックが見つかりません")
    }

    // 6. 値を読み取り
    buf := make([]byte, 255)
    n, err := chars[0].Read(buf)
    if err != nil {
        println("読み取り失敗 (通知専用の場合は正常):", err.Error())
    } else {
        println("読み取り値:", buf[:n])
    }

    // 7. 通知を受信
    //    (EnableNotificationsWithMode ではなく EnableNotifications を使用)
    must("enable notifications", chars[0].EnableNotifications(func(buf []byte) {
        println("通知受信:", buf)
    }))

    println("通知待ち受け中...")
    select {}
}

func must(action string, err error) {
    if err != nil {
        panic("failed to " + action + ": " + err.Error())
    }
}
```

このサンプルは以下のクロスプラットフォーム原則に従っています。

1. サービス UUID でデバイスを識別 (MAC アドレス非依存)
2. セントラル操作のみ使用
3. `EnableNotifications` で通知を受信
4. `Adapter.Address()` を使用しない
5. `ConnectionParams` はデフォルト値を使用
