# mefproxy v2.0

MCPE **1.1.0 – 1.21.90** multi-version RakNet proxy. Runs on `localhost:19132`, connects to any MCPE server.  
Clean module architecture — each cheat is a self-contained Go struct. No comments in source code.

---

## Changelog  v2.0

### Core
- **Multi-version support** — полная поддержка MCPE 1.1.0 – 1.21.90 (80+ версий, полная карта протоколов)
- **gameVersion spoof** — поле `gameVersion` в `config.json`; можно ввести любую строку версии (например `1.21.80`), и именно она будет передана серверу в `LoginPacket`, независимо от реальной версии MCPE клиента
- **Ghost bug fix** — исправлена утечка фантомных игроков: при трансфере/дисконнекте `PlayerList REMOVE` теперь корректно удаляет игроков через `uuidToEID` маппинг
- **State расширен** — `NearestPlayers(n)`, `PlayerByEID`, `PlayerByNick`, `PlayersInRange`, `PlayersInRangeSorted`, `PlayerCount`, `UpdatePlayerHealth`
- **API удвоен** — см. раздел API ниже

### Новые модули
- `TPAura` (Combat) — телепортируется к цели и немедленно атакует
- `Blink` (Misc) — буферизует движение, потом отправляет всё разом
- `PacketFlooder` (Network) — флудит сервер animate/move пакетами с заданным pps
- `AntiBot` (Network) — определяет ботов по нику, движению, позиции (энтропия, digit ratio, паттерны)
- `ConnectionSpoof` (Network) — подделывает DeviceOS, DeviceModel, InputMode при подключении
- `PacketRateLimiter` (Network) — ограничивает частоту пакетов клиента по типу, защита от самофлуда

### Команды (новые)
- `.version [ver]` — показать или установить спуф версии (`.version 1.21.80`, `.version off`)
- `.versions` — список всех 80+ поддерживаемых версий

### Улучшения модулей
- `KillAura` — настройка `targets` (до 5 целей), `rotate`; теперь атакует N ближайших
- `Nuker` — настройка `smart` (сортировка по дистанции от ближайшего к дальнему)
- `Flight` — синхронизирует `Speed` при включении
- `AutoSprint` — использует новый `proxy.SendSprint()`
- `Speed` — использует новый `proxy.SetMySpeed()`

---

## Структура проекта

```
mefproxy/
├── cmd/mefproxy/
│   ├── main.go             ← запуск, регистрация модулей, баннер
│   ├── commands.go         ← диспетчер команд
│   ├── cmd_modules.go      ← .toggle .modules .status .near .players
│   ├── cmd_server.go       ← .server .nick
│   ├── cmd_settings.go     ← .set .settings
│   └── cmd_version.go      ← .version .versions
│
├── proxy/
│   ├── proxy.go            ← ядро прокси, маршрутизация пакетов
│   ├── api.go              ← публичный API (40+ методов)
│   ├── config.go           ← Config с gameVersion
│   ├── versions.go         ← VersionProtocolMap 1.1→1.21.90
│   ├── state.go            ← GameState: players, EID, location
│   ├── hooks.go            ← реестр хуков
│   ├── client.go           ← локальный RakNet listener
│   └── server.go           ← коннектор к серверу
│
├── module/
│   ├── module.go           ← Module interface + Base struct
│   ├── category.go         ← Combat/Movement/Misc/Exploit/Visual/Network
│   ├── setting.go          ← Bool/Float/Int/Text/Slider settings
│   └── manager.go          ← Register, Toggle, ByCategory, StatusLine
│
├── modules/
│   ├── combat/
│   │   ├── hitbox.go
│   │   ├── killaura.go     ← multi-target, rotate
│   │   ├── criticals.go
│   │   ├── velocity.go
│   │   ├── reach.go
│   │   └── tpaura.go       ← NEW
│   ├── movement/
│   │   ├── nofall.go
│   │   ├── autosprint.go
│   │   ├── speed.go
│   │   ├── flight.go       ← speed sync
│   │   └── step.go
│   ├── misc/
│   │   ├── antiafk.go
│   │   ├── bypass.go
│   │   ├── chatlogger.go
│   │   ├── autoreconnect.go
│   │   ├── packetlogger.go
│   │   └── blink.go        ← NEW
│   ├── exploit/
│   │   ├── phase.go
│   │   ├── nuker.go        ← smart sort
│   │   ├── scaffold.go
│   │   ├── entityspeed.go
│   │   └── inventorymove.go
│   ├── visual/
│   │   ├── tracer.go
│   │   ├── esp.go
│   │   ├── breadcrumbs.go
│   │   ├── chestfinder.go
│   │   └── nametaghud.go
│   └── network/            ← NEW категория
│       ├── packetflooder.go
│       ├── antibot.go
│       ├── connectionspoof.go
│       └── packetratelimiter.go
│
├── pkg/
│   ├── packet/             ← MCPE protocol + encrypter/login
│   ├── transport/          ← RakNet
│   ├── nbt/                ← NBT codec
│   ├── world/              ← chunk/subchunk
│   ├── entity/             ← entity attributes
│   ├── block/ item/        ← block/item defs
│   ├── jose/               ← JWT/JWE crypto
│   ├── math/vec3/          ← 3D vector math
│   └── resource/           ← resource pack types
│
├── utils/                  ← buffer/network/skin/socks5/color helpers
└── config.json
```

---

## Build

```bash
cd cmd/mefproxy
go build -o mefproxy.exe .
```

---

## config.json

```json
{
  "address": "play.example.com:19132",
  "nick": "",
  "deviceModel": "XIAOMI MT84373Y48J",
  "inputMode": 2,
  "defaultInputMode": 0,
  "uiProfile": 0,
  "deviceOS": 0,
  "bypassRPDownload": false,
  "gameVersion": ""
}
```

| Поле | Описание |
|------|----------|
| `address` | IP:порт целевого сервера |
| `nick` | Спуф ника (пусто = ник аккаунта) |
| `deviceModel` | Модель устройства |
| `inputMode` | 1=mouse 2=touch 3=controller |
| `deviceOS` | 1=android 2=ios 7=win10 |
| `bypassRPDownload` | Пропустить загрузку ресурспаков |
| `gameVersion` | **Спуф версии** — любая строка, например `1.21.80`. Пусто = версия клиента. Сервер увидит именно эту версию. |

### Примеры gameVersion

```json
"gameVersion": "1.21.80"
"gameVersion": "1.20.0"
"gameVersion": "1.8.0"
"gameVersion": "2.0.0"
```

Можно ввести абсолютно любую строку — прокси передаст её в `ClientData.GameVersion`. Если версия есть в `VersionProtocolMap`, команда `.version` покажет протокол.

---

## Команды в игре

| Команда | Описание |
|---------|----------|
| `.help` | Список команд |
| `.modules` | Все модули по категориям |
| `.toggle <модуль>` | Включить/выключить |
| `.set <мод> <пар> <знач>` | Изменить настройку |
| `.settings <модуль>` | Список настроек модуля |
| `.status` | Статус прокси |
| `.players` | Игроки рядом |
| `.near` | Ближайший игрок |
| `.server <ip:port>` | Переключить сервер |
| `.nick <ник>` | Сменить ник |
| `.version` | Показать текущий спуф версии |
| `.version <ver>` | Установить спуф (напр. `1.21.80`) |
| `.version off` | Выключить спуф |
| `.versions` | Список всех 80+ версий |

---

## Написание модуля

```go
package network

import (
    "mefproxy/module"
    "mefproxy/pkg/packet"
    "mefproxy/proxy"
)

type MyModule struct {
    module.Base
}

func NewMyModule() *MyModule {
    m := &MyModule{
        Base: module.NewBase("MyModule", "Описание", module.Network),
    }
    m.AddSetting(module.NewSlider("range", 4.0, 1.0, 10.0, 0.5))
    m.AddSetting(module.NewBool("verbose", false))
    return m
}

func (m *MyModule) Init(p *proxy.Proxy) { m.Base.Init(p) }

func (m *MyModule) Enable() {
    m.Base.Enable()
    m.Proxy.OnServerPacket(m.Name(), func(pk []byte) bool {
        if pk[0] != packet.IDAddPlayerPacket {
            return false
        }
        // return true = DROP (блокировать)
        // return false = PASS (пропустить)
        return false
    })
}

func (m *MyModule) Disable() {
    m.Base.Disable()
    m.Proxy.RemoveHooks(m.Name())
}
```

Зарегистрируй в `cmd/mefproxy/main.go`:
```go
network.NewMyModule(),
```

---

## API proxy.Proxy (.0)

### Отправка пакетов
```go
p.SendToClient(pk)           // отправить пакет клиенту
p.SendToServer(pk)           // отправить пакет серверу
p.SendRawToClient(raw)       // raw байты клиенту
p.SendRawToServer(raw)       // raw байты серверу
p.SendBatchToClient([]pk)    // батч пакетов клиенту
p.SendBatchToServer([]pk)    // батч пакетов серверу
```

### Уведомления
```go
p.Notify(msg)
p.NotifyF(format, args...)
p.NotifyTitle(title, sub, fadeIn, stay, fadeOut)
p.NotifyActionBar(msg)
p.DelayedNotify(msg, delay)
```

### Состояние
```go
p.MyEID()
p.MyLocation()
p.Players()
p.NearestPlayer()
p.NearestPlayers(n)
p.PlayerByEID(eid)
p.PlayerByNick(nick)
p.PlayerCount()
p.PlayersInRange(radius)
p.PlayersInRangeSorted(radius)
p.IsClientOnline()
p.IsServerOnline()
p.IsFullyOnline()
```

### Действия
```go
p.AttackEntity(eid)
p.SwingArm()
p.SendMove(x, y, z, yaw, pitch, onGround)
p.SendMoveAtCurrentPos(onGround)
p.BreakBlock(x, y, z, face)
p.PlaceBlock(x, y, z, face)
p.SetMotion(eid, mx, my, mz)
p.SendSprint(bool)
p.SetMySpeed(val)
p.SetMyHealth(val)
p.SetMyFood(val)
p.SetAdventureFlags(allowFlight, isFlying)
p.SetEntityData(eid, meta)
p.SetPlayerHitbox(eid, w, h)
```

### Хуки
```go
p.OnClientPacket(key, fn)
p.OnServerPacket(key, fn)
p.OnConnect(key, fn)
p.OnDisconnect(key, fn)
p.OnServerDisconnect(key, fn)
p.RemoveHooks(key)
```

### Конфиг
```go
p.GetConfig()
p.SetConfig(fn)
p.SwitchServer(addr)
p.GetSpoofedVersion()
p.SetSpoofVersion(ver)
p.GetVersionInfo()          // → (version, protocol, found)
```

---

## Версии и протоколы

```
1.1.0  → 113    1.8.0  → 313    1.16.0   → 407
1.2.0  → 136    1.9.0  → 332    1.17.0   → 440
1.4.0  → 261    1.10.0 → 340    1.18.0   → 475
1.5.0  → 274    1.11.0 → 354    1.19.0   → 527
1.6.0  → 282    1.12.0 → 361    1.20.0   → 589
1.7.0  → 291    1.13.0 → 388    1.21.0   → 685
                1.14.0 → 389    1.21.90  → 819
```

Полный список: `.versions` в игре.
