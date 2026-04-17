# Proxy Backend Extraction Plan

Статус: completed implementation  
Дата: 2026-04-17  
Worktree: `/Users/belief/dev/projects/network_debugger_proxy_pkg_plan`  
Ветка: `refactor/proxy-core-extract`

## Цель

Из текущего backend нужно получить набор reusable Go packages, которые:

1. полезны вне `network-debugger`
2. не содержат Flutter/frontend-специфики
3. имеют узкие и понятные package boundaries
4. выдерживают рост функциональности без повторного сваливания transport, capture, API и storage в один слой
5. сохраняют работоспособность текущего приложения через adapters и compatibility endpoints

## Финальный результат

Extraction доведён до конца по целевому состоянию:

- `proxykit` вынесен в отдельный dedicated repo workspace: `/Users/belief/dev/projects/proxykit`
- `network-debugger` использует внешний модуль `github.com/777genius/proxykit`
- public transport foundation больше не зависит от Flutter/frontend contracts
- app-specific compatibility surface сохранён в adapters и binary composition layer
- полный Go verification после split снова зелёный
- обязательные Flutter/Dart smoke и real frontend web build тоже зелёные

Ключевые закрытые этапы:

1. `scripts` monolith добит до отдельных seams:
   - `script_test_service`
   - `script_project_service`
   - `script_archive_service`
   - `script_examples_provider`
2. `monitor` delivery policy вынесена в отдельный adapter seam:
   - raw monitor WS и Socket.IO сохранены как разные каналы
   - `MonitorHub` оставлен dumb broadcaster-ом
3. binary/app composition очищена:
   - embedded web mounting собран в reusable composition helper
   - `network-debugger-web` и `wsapp` больше не держат свои копии SPA wiring
4. `proxykit` hardened до community-grade surface:
   - compile-able examples
   - self-contained package tests
   - dedicated `LICENSE`
   - отдельный repo workspace
5. verification закрыт end-to-end:
   - `go test ./... -count=1` по root module
   - `go test ./...` и `go test -race ./...` по `proxykit`
   - `flutter test`
   - Dart package suites
   - `make frontend-build-web`
   - compile pass для `cmd/network-debugger-web` уже с реальными built assets

## Что уже сделано

### Реализован public module

Создан отдельный модуль:

- `github.com/777genius/proxykit`

Причина выбора именно такого module path:

- не тащит в public imports слово `flutter`
- не цементирует app-specific идентичность текущего репозитория
- позволяет локально развивать модуль через `replace`, а потом вынести в отдельный repo почти без изменений

Финальная форма на этом этапе:

- модуль уже вынесен в отдельный repo workspace: `/Users/belief/dev/projects/proxykit`
- текущее приложение подключает его как внешний sibling module, а не как nested incubation directory
- это совпадает с рекомендуемым направлением из официальных Go материалов: отдельный чистый repo под отдельный module проще для versioning, tagging и community adoption

Ориентиры:

- [Managing module source](https://go.dev/doc/modules/managing-source)
- [Go Wiki - Multi-Module Repositories](https://go.dev/wiki/Modules)
- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Package names](https://go.dev/blog/package-names)
- [Keeping Your Modules Compatible](https://go.dev/blog/module-compatibility)

### Уже вынесенные public пакеты

- `proxykit/proxyruntime`
  - lifecycle для forward proxy listener
  - lifecycle для SOCKS5 listener
  - apply/restart semantics
  - port conflict diagnostics
- `proxykit/socketio`
  - парсинг Socket.IO event-style пакетов
- `proxykit/proxyhttp`
  - `http.Transport` policy для proxy transport
  - hop-by-hop header stripping
  - WebSocket upgrade detection
  - absolute-URI detection
- `proxykit/wsproxy`
  - reusable WebSocket transport engine
  - hook-driven session and frame observation
  - query target resolver helper
  - optional plaintext fallback for TLS-mismatch upstreams

### Уже выполненная интеграция в приложение

- `internal/infrastructure/httpapi/router.go` использует public `proxyruntime`
- `internal/infrastructure/httpapi/proxy_api.go` использует public `proxyruntime`
- `internal/infrastructure/httpapi/ws_sio.go` использует public `socketio`
- `internal/infrastructure/httpapi/httpproxy_unified.go` использует public `proxyhttp`
- `internal/infrastructure/httpapi/httpproxy.go` использует public `reverse` как transport engine и public `cookies`/`proxyhttp` как shared policies
- `internal/infrastructure/httpapi/forwardproxy.go` использует public `proxyhttp.IsAbsoluteURL`
- `internal/infrastructure/httpapi/forwardproxy.go` использует public `forward` и `connect`
- `internal/infrastructure/httpapi/mitm.go` использует public `mitm`
- `internal/infrastructure/httpapi/wsproxy.go` использует public `wsproxy`
- старые internal пакеты пока оставлены как compatibility shims для безопасной миграции тестов

### Extraction статус

- `reverse HTTP transport` уже переведён с app-owned `httputil.ReverseProxy` на public `proxykit/reverse` через adapter flow
- `forward HTTP`, `CONNECT`, `MITM`, `WS`, `Socket.IO`, `cookies`, `runtime` уже имеют public-safe seams
- `sessionV1/httpMeta/sizes` projection больше не размазана по REST и Socket.IO list/init paths
- для session projection теперь есть отдельный internal seam, который переиспользуется и в `/_api/v1/sessions`, и в realtime session list delivery
- realtime fanout/init/detail-room catch-up больше не живут прямо в `socketio_sessions.go` как transport-owned монолит, а вынесены в отдельный internal orchestration seam
- legacy/v1 session detail loading, frame body semantics и legacy event aliasing тоже больше не размазаны по handler switch-блокам, а вынесены в отдельный internal detail seam
- session list query parsing и `sessions_cleared` lifecycle тоже собраны в отдельные internal seams, а не размазаны по legacy/v1/capture handlers
- HAR export/import orchestration тоже больше не живёт в handler'ах, а вынесена в отдельный internal service seam с сохранением app compatibility routes и summary payloads
- raw monitor WebSocket delivery тоже больше не живёт внутри `MonitorHub` как смешанная bus+socket реализация, а вынесена в отдельный delivery seam
- `/_api/v1/proxy/config` больше не держит load/validate/merge/apply прямо в handler'е, а собран в отдельный admin service seam
- `/_api/v1/throttle` и `/_api/v1/throttle/profiles` больше не держат runtime overlay, persistence и profile CRUD прямо в handler'ах, а вынесены в отдельный admin service seam
- `/_api/v1/settings` тоже больше не держит parsing/apply/persistence прямо в handler'е, а собран в отдельный admin service seam
- `/_api/v1/settings` теперь ещё и умеет корректный partial update semantics: theme/font updates больше не должны неявно сбрасывать `responseDelay`
- dual-theme compatibility для frontend тоже закрыта на backend стороне: `highlightThemeLight` и `highlightThemeDark` теперь имеют persistence и GET/POST поддержку без потери legacy `highlightTheme`
- `/_api/v1/tags/*` и session tag/annotation endpoints больше не держат validation, bulk policy и storage orchestration прямо в handler'ах, а вынесены в отдельный admin service seam
- `/_api/v1/process/*` тоже больше не держат config/helper orchestration прямо в handler'ах, а собраны в отдельный admin service seam
- `tags` и `process` endpoints теперь ещё и безопаснее для headless wiring: вместо nil-driven паники они умеют возвращать `503 FEATURE_DISABLED`
- `/_api/v1/mapping/*` больше не держат config overlay, rules reorder/upsert/delete, runtime refresh и upload/spool policy прямо в handler'ах, а собраны в отдельный admin service seam
- `/_api/v1/intercept/*` тоже больше не держат rules/config CRUD, pending queue parsing, item path decoding и request/response decision validation прямо в handler'ах, а вынесены в отдельный admin service seam
- core `/_api/v1/scripts` CRUD/toggle flow тоже больше не держит create/update/toggle orchestration прямо в handler'ах, а собран в отдельный admin service seam с сохранением compile-on-create, recompilation invalidation и enable-policy semantics
- первоначальный transport/app monolith разрезан до public transport packages и app adapters
- оставшиеся app-specific compatibility policies собраны в отдельные internal services и composition seams
- скрытой крупной business orchestration внутри transport handlers больше не осталось

## Реальная карта сцепок после аудита кода

### Главный узел проблемы

Самая тяжёлая сцепка сосредоточена в `internal/infrastructure/httpapi`.

Там сейчас смешаны:

- HTTP reverse proxy
- HTTP forward proxy
- CONNECT tunneling
- MITM
- WebSocket proxy
- Socket.IO decoding
- session/frame/event capture
- HTTP transaction persistence
- monitor broadcasting
- runtime wiring
- mapping
- intercept
- scripting
- cookie rewriting
- spool files
- frontend compatibility endpoints

Это и есть центральный монолит, который надо распилить по причинам изменения, а не по типам файлов.

### Самые тонкие участки

#### 1. `handleHTTPProxy`

Это не просто reverse proxy handler. Он одновременно отвечает за:

- target resolution через `_target`
- path/query merge semantics
- cookie isolation policies
- request/response scripting
- request/response interception
- local/remote mapping
- preview generation
- body spooling
- HTTP transaction persistence
- frontend monitor events

Вывод:

- transport engine нельзя вытаскивать как копию этого handler
- сначала нужен hook-driven reverse engine
- mapping/script/intercept/capture должны остаться extension layer, а не стать частью core API

#### 2. `handleForwardProxy`

Здесь смешаны:

- regular forward HTTP proxy
- CONNECT tunnel
- forward WebSocket upgrade
- MITM entry point
- request/response logging
- mapping/intercept reuse

Вывод:

- forward package надо строить как composition из:
  - forward HTTP engine
  - CONNECT tunneling
  - optional WS-over-forward bridge
  - optional MITM package

#### 3. `handleWSProxy`

Здесь смешаны:

- target normalization `http(s) -> ws(s)`
- query merge semantics
- session lifecycle
- monitor events
- live session registry
- WebSocket piping
- Socket.IO event decoding
- fallback `wss -> ws` при TLS mismatch

Вывод:

- это лучший следующий extraction seam после уже вынесенных low-level пакетов
- raw WS engine должен жить отдельно от SessionService и MonitorHub
- Socket.IO decoding должен быть observer capability, а не обязательной частью WS transport

#### 4. Cookie rewriting

`cookies.go` уже содержит сложную policy-логику:

- `off/auto/isolate`
- namespace hashing
- domain/path strategy
- secure/samesite нюансы

Вывод:

- это отдельная ответственность
- переносить в core только после выделения собственного policy package
- не вшивать cookie strategy прямо в `reverse` engine

#### 5. Frontend и Dart SDK contracts реально внешние

Не только frontend приложения, но и опубликованные Dart helper packages уже завязаны на:

- `/httpproxy`
- `/wsproxy`
- `/_api/v1/proxy/config`
- `/healthz`
- `/_health`
- отдельный proxy port lifecycle
- CLI flags `--api-port`, `--proxy-port`, `--data-dir`
- поведение reverse helpers из `http_debugger` и `dio_debugger`
- поведение reverse/forward helpers из `web_socket_debugger`, `web_socket_channel_debugger`, `socket_io_debugger`
- env/define contracts:
  - `SOCKET_PROXY`
  - `SOCKET_PROXY_PATH`
  - `SOCKET_PROXY_MODE`
  - `SOCKET_PROXY_ENABLED`
  - `SOCKET_PROXY_ALLOW_BAD_CERTS`
  - `SOCKET_UPSTREAM_URL`
  - `SOCKET_UPSTREAM_TARGET`
  - `SOCKET_UPSTREAM_PATH`
  - `DIO_DEBUGGER_ENABLED`

Вывод:

- это не internal detail
- это adapter boundary
- менять можно только при наличии compatibility layer и обновления frontend/Dart packages сразу в том же цикле

#### 6. Mapping/intercept/script - это не один "middleware layer"

После дополнительного аудита видно, что три похожие на вид capability на деле имеют разные причины изменения:

- mapping runtime:
  - chainable remote rewrites
  - `StopProcessing`
  - `PreserveHost`
  - локальные file/blob responses
- intercept:
  - queue
  - timeout
  - overflow strategy
  - pending items
  - explicit continue/cancel lifecycle
- scripting:
  - runtime executors
  - validation
  - body/context truncation limits
  - plugin cache invalidation

Вывод:

- их нельзя прятать под один public `middleware` abstraction
- transport engine должен давать mutation/observation seams
- сами mapping/intercept/script orchestration должны оставаться app-level extensions до отдельного зрелого extraction

#### 7. Runtime config и launcher flow - это отдельный actor

`/_api/v1/proxy/config` и launcher behavior живут своей жизнью:

- DTO shape для UI
- validation ports/auth mode/conflicts
- сохранение runtime config
- `Apply(...)` к listener runtime
- UX expectations settings page

Вывод:

- `proxyruntime` должен оставаться headless lifecycle package
- REST DTO, validation policy и settings UX остаются в приложении
- public module не должен знать про этот endpoint или его JSON shape

#### 8. Body spool lifecycle живёт дольше transport request

Слой stored bodies завязан сразу на несколько акторов:

- transport capture
- session details UI
- `frames/{id}/body`
- HAR export/import
- session deletion
- clear-all
- imported session cleanup
- background spool GC

Вывод:

- это не часть raw transport API
- до появления отдельного reusable body-store abstraction этот lifecycle должен оставаться app concern
- `observe` может давать body handles/references, но не владеть temp-file policy

#### 9. App monitor/live-session layer - отдельный actor, не часть public transport

Текущий app layer поверх transport делает намного больше, чем просто слушает события:

- monitor bus
- Socket.IO subscriptions
- quick filters
- aggregate recalculation
- session-specific rooms
- live connection registry для WS
- app-shaped `sessions:init/upsert/remove`

Вывод:

- это не часть `observe` public API
- `observe` должен давать достаточно нейтральных signals для adapters
- aggregate/filter/room semantics должны остаться в приложении

#### 10. HAR и imported sessions - отдельная migration boundary

HAR support уже привязан к нескольким app-specific решениям:

- imported sessions имеют `ClientAddr == "imported"`
- `replace_all` и `replace_imported`
- attach WebSocket messages к последнему HTTP upgrade entry
- include-bodies зависит от current spool/body file policy

Вывод:

- HAR import/export нельзя считать просто ещё одним transport feature
- это отдельный adapter/application concern
- extraction transport layers не должен ломать imported-session semantics и body ownership rules

#### 11. Session classification/filtering - это derived app model

После дополнительного прохода по tests и helper-ам видно, что sessions UI работает не только по raw transport data, а по derived classification:

- status groups `1xx..5xx`
- type tags `http/ws/graphql/json/form/xml/js/css/media/document`
- GraphQL deep scan по body only when explicitly requested
- capture scope semantics
- cache/CORS/preflight derived metadata inside `httpMeta`

Вывод:

- это не transport concern
- это app-layer projection над observed data
- public extraction не должен цементировать эти derived tags/filters в core packages

#### 12. Published package docs уже обещают configurable mounting semantics

Документация Dart packages уже обещает пользователям:

- reverse mode by default
- forward mode as explicit alternative
- configurable `proxyHttpPath` / `proxyPath`
- возможность unified endpoint mounting вроде `/proxy`
- auto mode and fallback behavior in некоторых helpers

Вывод:

- public Go packages должны проектироваться как path-agnostic composable handlers
- hardcoded `/httpproxy` и `/wsproxy` остаются app compatibility defaults, а не фундамент public API
- examples и docs будущего Go module должны показывать composable mounting, а не только current app routes

#### 13. Analytics/performance пока завязаны на app-shaped approximations

Performance слой сейчас использует:

- `Session.StartedAt/ClosedAt` как approximation duration
- `Session.Target` как endpoint key
- отсутствие стабильной public observe model для richer analytics

Вывод:

- analytics/performance пока нельзя делать частью extraction foundation
- сначала нужен stable observe model, потом уже reusable analytics packages, если они реально нужны сообществу

#### 14. Realtime inspector protocol - это отдельный app contract, не generic observe API

После дополнительного аудита frontend/backend realtime видно, что UI живёт не на "сырых transport events", а на своём поверхностном protocol:

- Socket.IO path `/_api/v1/monitor/io/socket.io/`
- connection bootstrap с auto-init
- `sessions:subscribe`
- `sessions:init`
- `sessions:upsert`
- `sessions:remove`
- `aggregate:update`
- detail-page rooms через:
  - `session:subscribe`
  - `session:unsubscribe`
  - `session:frames`
  - `session:events`
  - `session:http`
- REST fallback при `connect_error`/`disconnect`
- debounce/recompute aggregates per subscriber
- drop-on-overflow semantics в `MonitorHub` и Socket.IO adapters

Вывод:

- это не public `observe` contract
- это app-level projection/delivery protocol
- `observe` должен давать сырьё для такого protocol, но не тащить в себя event names, rooms, aggregate debounce и reconnect fallback policy

#### 15. Не все sessions рождаются из proxy transport

Дополнительный аудит `ComposeService` показал, что часть session model создаётся synthetic путём:

- compose создаёт `Session`
- добавляет synthetic request frame
- делает direct `http.Client` request
- пишет `HTTPTransaction`
- добавляет response frame
- шлёт те же monitor events, что и proxy-based flows

Вывод:

- app session model шире, чем просто proxy observation
- нельзя проектировать public `observe` так, будто все session/frame/http_tx события обязаны происходить из reverse/forward/ws engines
- synthetic producers вроде compose должны остаться app concern или отдельным higher-level package, но не диктовать shape core transport API

#### 16. Route surface зависит от конкретного binary

Аудит `cmd/network-debugger-web` и `cmd/wsapp` показал, что не все binaries публикуют одинаковые пути:

- embedded-web binaries монтируют:
  - `/_api/`
  - `/api/`
  - `/httpproxy`
  - `/_ws`
  - SPA static fallback на `/`
- при этом они не обязаны публиковать тот же path set, что plain backend/proxy binary

Вывод:

- route surface - это app/binary composition layer
- public core packages должны быть mount-path agnostic
- extraction не должен случайно приравнять текущие binary-specific aliases к public package defaults

#### 17. TLS split уже выражает архитектурную границу

В `cmd/network-debugger` и `cmd/wsproxy` уже есть важное разделение:

- plain HTTP server держит forward proxy / CONNECT semantics
- optional TLS server использует `NewRouterWithoutForwardProxy`
- reverse/REST под TLS могут получить HTTP/2
- CONNECT intentionally не переносится на TLS listener

Вывод:

- это не implementation detail, а уже существующая transport boundary
- `forward` и `reverse` extraction должны уважать split "forward proxy listener" vs "reverse/REST listener"
- нельзя случайно собрать один public handler, который размоет этот boundary и испортит CONNECT or h2 story

#### 18. Capture lifecycle - это app visibility model, не transport capability

Дополнительный аудит capture/storage/frontend показал, что recording state влияет не только на запись, но и на видимость данных:

- memory store автоматически присваивает `CaptureID` только когда recording включён
- paused sessions остаются `CaptureID == nil`
- UI и REST используют:
  - `captureId=current`
  - `captures=all`
  - `includeUnassigned=true`
  - `captureScope=current|all`
  - `pausedSince`
- realtime layer дополнительно отбрасывает "слишком новые" unassigned sessions на клиенте, если recording уже paused
- `_resetCapture=true` на `/httpproxy` и `POST /_api/v1/capture/reset` запускают cleanup + close live sessions + broadcast `sessions_cleared` + start new capture

Вывод:

- capture/recording сейчас уже часть app orchestration and visibility model
- это нельзя вшивать в public transport or observe API как будто это универсальная proxy capability
- public foundation может дать hooks, но current capture semantics, reset UX и paused visibility policy должны остаться в app layer

#### 19. Synthetic ingest уже является внешним contract

Помимо compose и HAR/import в системе уже есть отдельный published-style ingest producer:

- `POST /_api/v1/ingest/firebase_database`
- session kind `firebase_database`
- explicit `captureId=current|N`
- frame preview/body sanitization, size limits и spool policy
- auth policy через `AdminToken` или loopback/private-network allowance
- frontend уже знает про этот kind и отдельные filters/badges

Вывод:

- app session graph уже включает protocol families, которые не идут через reverse/forward/wsproxy
- public proxy packages не должны внезапно стать host для ingest DTO/auth/capture semantics
- synthetic producers должны оставаться app concern или отдельными protocol-specific adapters

#### 20. Admin token - это shared app security boundary

Дополнительный аудит показал, что `X-Admin-Token` уже не локальная настройка одного endpoint:

- mapping API
- intercept API
- firebase ingest API
- frontend settings/prefs and several feature clients already знают этот header
- loopback/private-network bypass policy тоже сейчас app-specific и не едина для всех endpoints

Вывод:

- auth around admin/debugging endpoints - это app security layer
- public `proxykit` core не должен знать про `AdminToken`, `X-Admin-Token`, loopback bypass или remote ingest policy
- extraction должен удержать transport packages headless, а admin tooling/security policy оставить на app edge

#### 21. У приложения уже два разных monitor channels с разной семантикой

Дополнительный аудит frontend показал, что realtime delivery уже split by purpose:

- Socket.IO channel `/_api/v1/monitor/io/socket.io/`
  - sessions list snapshots
  - aggregate updates
  - session detail rooms
- raw WebSocket channel `/_api/v1/monitor/ws`
  - `sessions_cleared`
  - `session_error`
  - `intercept_*` refresh triggers

Вывод:

- это не один generic event bus
- у каналов уже разная delivery semantics и разный consumer set
- public `observe` не должен фиксировать current channel topology приложения

#### 22. Error taxonomy уже участвует в UI behavior, а не только в логах

Дополнительный аудит ошибок показал, что `classifyProxyError` уже влияет на several app behaviors:

- `session_error` monitor payload несёт:
  - `category`
  - `code`
  - `userMessage`
  - `message`
  - `raw`
  - `target`
  - `method`
- `httpMeta` дополнительно проецирует:
  - `errorCategory`
  - `errorCode`
  - `errorUserMessage`
  - `errorMessage`
- UI filters and badges уже завязаны на codes like:
  - `CANCELED`
  - `TIMEOUT`
  - `DNS_ERROR`
  - `TLS_ERROR`
  - `SERVER_UNAVAILABLE`
- cancelled requests intentionally скрываются в notifications and some filters
- WS errors обрабатываются отдельно от HTTP notifications

Вывод:

- error taxonomy сейчас уже часть app projection and UX policy
- public transport hooks могут отдавать raw errors, но stable UI codes/messages не надо цементировать в core prematurely
- classification, suppression and notification policy должны остаться в app layer

#### 23. Runtime simulation уже стала app-level policy, а не просто transport tweak

Дополнительный аудит throttling/settings/frontend показал, что "симуляция сети" уже живёт как отдельная пользовательская capability:

- persisted runtime state для:
  - `ResponseDelayMs`
  - `ThrottleEnabled`
  - `ThrottleDownKbps`
  - `ThrottleUpKbps`
  - `ThrottlePacketLoss`
  - `ThrottleLatencyMs`
  - `ThrottleLatencyJitter`
  - `ThrottleOffline`
- отдельные REST surfaces:
  - `/_api/v1/throttle`
  - `/_api/v1/throttle/profiles`
- profiles CRUD с user-facing preset semantics
- одна и та же policy применяется сразу в:
  - reverse HTTP
  - forward HTTP/CONNECT
  - WebSocket piping
- frontend settings UI уже трактует offline/delay/throttle как persisted product features, а не как debug-only knobs

Вывод:

- runtime simulation нельзя тихо втащить в public `reverse` или `forward` как "ещё пару опций"
- это отдельный actor со своей lifecycle/config/profile semantics
- для community-grade extraction это либо app-layer policy, либо позже отдельный optional package, но не часть transport foundation v1

#### 24. Manual live WS control - это app control surface, не generic ws transport

Дополнительный аудит `LiveSessions`, router aliases и manual send flow показал, что приложение умеет не только проксировать WS, но и активно вмешиваться в живые сессии:

- registry активных WS соединений живёт отдельно от storage
- route `/api/sessions/{id}/ws/send` уже часть legacy/app surface
- manual send использует direction semantics:
  - `client->upstream`
  - `upstream->client`
- reset/cleanup flows уже зависят от способности закрыть все live connections

Вывод:

- live manual send/control - это higher-level app debugging feature
- public `wsproxy` engine должен оставаться headless transport/observation layer
- registry, command routes, send-text policy и admin UX нельзя цементировать в public core

#### 25. Process enrichment и helper tooling - это локальная machine-integration boundary

Дополнительный аудит process detection показал, что здесь уже не один маленький helper, а отдельный набор app responsibilities:

- session creation может автоматически enrich-ить `ProcessInfo` по `ClientAddr`
- существует persisted config surface:
  - `/_api/v1/process/config`
  - `/_api/v1/process/helper/status`
  - `/_api/v1/process/helper/install`
- helper install/status/version и cache/fallback semantics уже часть settings UX
- реализация зависит от platform-specific detectors, helper binary discovery, icon extraction и local machine permissions

Вывод:

- process enrichment нельзя считать частью reusable proxy transport foundation
- public packages могут когда-нибудь дать neutral extension seam для session enrichment, но не должны тянуть helper install/status/config в core
- current process model, helper lifecycle и local-machine policy остаются app layer

#### 26. Tags/annotations - это app metadata model поверх sessions

Дополнительный аудит tags backend/frontend показал, что tags уже участвуют не только в details UI, но и в фильтрации и realtime payloads:

- REST/API surface уже включает:
  - `/_api/v1/tags/predefined`
  - `/_api/v1/tags/bulk`
  - `/_api/v1/sessions/{id}/tags`
  - `/_api/v1/sessions/{id}/annotations`
- есть predefined tags, session tags, bulk add/remove и per-session annotations
- frontend хранит selected tags в prefs и использует их как часть filters/realtime subscribe payload
- Dart client уже содержит fallback behavior между session-level tags API и bulk API

Вывод:

- tags/annotations - это не reusable transport concern
- это app metadata, filtering и collaboration layer поверх session model
- public foundation должен давать raw observations, а tags/annotations и related REST/prefs/filter semantics должны остаться в приложении

#### 27. Structured HTTP preview schema уже стала отдельным UI protocol

Дополнительный аудит `httpproxy.go`, preview tests и frontend HTTP inspector показал, что `Frame.Preview` для HTTP давно используется не как "кусок текста", а как структурированная схема:

- request preview несёт:
  - `type=http_request`
  - `method`
  - `url`
  - masked `headers`
  - optional `headersRaw`
  - `body`
  - `bodyRawSize`
  - `form` preview for urlencoded/multipart
- response preview несёт:
  - `type=http_response`
  - `status`
  - masked `headers`
  - optional `headersRaw`
  - `body`
  - `bodyRawSize`
  - `cookieSummary`
  - `tls`
  - `timings`
- frontend и compose уже парсят эту схему для:
  - truncation detection
  - copy-as-cURL
  - Open in Compose
  - AI copy
  - request/response info tabs
  - `httpMeta` warmup fallback

Вывод:

- preview schema - это уже app-facing protocol, а не raw transport observation
- public core не должен стабилизировать текущий JSON preview shape как свой foundation API
- правильная граница - raw observations в `observe`, а current preview JSON contract собирает app adapter

#### 28. `frames/{id}/body` - это полноценный compatibility contract, а не storage helper

Дополнительный аудит body-serving endpoint и frontend body panel показал, что здесь уже есть зафиксированная UX-семантика:

- `/_api/v1/sessions/{id}/frames/{frameId}/body` возвращает:
  - raw bytes from file
  - или preview fallback, если `BodyFile` отсутствует
- response headers уже значимы для UI:
  - `X-Body-Source=file`
  - `X-Body-Source=preview`
  - `X-Frame-Id`
- frontend различает несколько semantic outcomes:
  - preview fallback when body capture disabled
  - `404` frame not found
  - `410 BODY_FILE_EXPIRED` after GC/delete
  - `413 BODY_TOO_LARGE` over 100 MB
  - normal raw body success

Вывод:

- это не просто "прочитать файл из spool"
- body-serving contract уже часть inspector UX и fallback logic
- extraction body/capture layers не должен ломать status/header semantics этого endpoint

#### 29. `/_api/v1/sessions` уже является server-side projection protocol

Дополнительный аудит repository/UI/tests показал, что список сессий давно перестал быть "сырым list endpoint":

- server-side filters уже включают:
  - `q`
  - `_target`
  - `tags`
  - `types`
  - `status`
  - `scan=graphql`
  - `captureId=current|N`
  - `captures=all`
  - `includeUnassigned`
  - `limit`
- response items уже обогащаются:
  - `httpMeta`
  - `sizes`
  - `processInfo`
  - derived counters used for `isSocketIo`
- frontend stores, quick filters, nd_test_support и e2e tests уже завязаны именно на этот projected surface, а не на raw sessions only

Вывод:

- `/_api/v1/sessions` - это app query/projection layer
- public `observe` и transport packages не должны утянуть в себя types/status/tags/graphql-scan/filter semantics
- app adapters должны сохранить этот list protocol как compatibility surface

#### 30. Proxy runtime config уже содержит отдельный privacy/UX contract

Дополнительный аудит `proxy_api.go`, tests и settings/integrations UI показал, что `/_api/v1/proxy/config` уже несёт не только runtime, но и deliberate app semantics:

- GET intentionally не возвращает SOCKS `user/pass`
- POST валидирует:
  - port range
  - `authMode`
  - forward-vs-socks port conflict
- по коду сохранение credentials работает как partial update:
  - non-empty `user/pass` overwrite stored values
  - empty values не используются как explicit clear operation
- integrations UI читает этот endpoint как source of truth для forward proxy port, но не для secrets

Вывод:

- это уже отдельный app security/privacy contract, а не просто `ApplyConfig`
- public `proxyruntime` не должен знать про credential redaction, partial-update semantics или settings UX assumptions

## Зафиксированные внешние инварианты

Это не "желательно сохранить", а реальные контракты, которые уже зашиты в frontend, Dart packages, e2e tests и UX.

### Runtime и startup

- API health должен отвечать и на `/healthz`, и на `/_health`
- `readyz` тоже уже существует как compatibility/readiness endpoint и сейчас возвращает именно `ready`, а не `ok`
- proxy listener должен реально подниматься на отдельном порту, а не логически существовать только в API process
- CLI contract `--api-port`, `--proxy-port`, `--data-dir` нельзя ломать до появления нового совместимого launcher layer
- root `/` у backend binary сейчас намеренно возвращает informational HTML landing page, а не `404`, чтобы launcher/browser UX не выглядел "сломавшимся"
- browser auto-open у Go binary сейчас строго opt-in:
  - `--open-browser` или `OPEN_BROWSER=1`
  - `NO_BROWSER=1` и `-cli` отключают auto-open
  - Dart launchers и desktop app уже сами прокидывают `NO_BROWSER=1`, а package launcher дополнительно использует `DEV_MODE=1`
- значит browser/startup UX - это app/launcher concern, а не часть public proxy foundation
- desktop frontend сейчас использует и health check, и log-based readiness hints вроде `started`, `listening`, `ready`
- startup order сейчас тоже важен:
  - settings overlay накладывается на env-config
  - затем поднимается runtime
  - затем стартует spool GC

### Proxy port multiplexing contract

На proxy port текущая система уже держит не только forward proxy:

- reverse HTTP endpoint `/httpproxy`
- reverse WS endpoint `/wsproxy`
- forward HTTP/CONNECT proxy
- health endpoints на proxy listener
- причём health short-circuit на proxy listener должен срабатывать только для non-proxy-style requests
- absolute-URI requests и `CONNECT` нельзя случайно перехватить как local `/healthz`, иначе сломается обычный forward proxy на upstream health URLs

Вывод:

- для текущего приложения proxy port это multiplexed app endpoint
- public packages могут давать отдельные engines, но app adapter обязан сохранить совместимую multiplexing composition

### Reverse HTTP contract

- `/httpproxy` и `/httpproxy/` должны обе работать
- legacy alias `/_api/v1/httpproxy` тоже уже зафиксирован tests и routing-ом
- `_target` остаётся ключевым compatibility parameter
- repeated query keys, unicode и пробелы должны доходить до upstream корректно
- query из `_target` и query входящего запроса должны мерджиться, при этом входящие параметры имеют приоритет
- redirect chain должна продолжать работать через proxy
- уже proxy-URL нельзя перепроксировать второй раз
- UI умеет через `/httpproxy` повторять исходный HTTP request с оригинальным методом, headers и body
- published Dart reverse helpers должны сохранять bypass semantics:
  - `skipPaths/skipHosts/skipMethods`
  - `allowPaths/allowHosts/allowMethods`
  - merged `upstreamBaseUrl + original URL` без потери path/query semantics
- published docs уже подразумевают configurable path mounting, включая возможный unified `/proxy`

### Route aliases и legacy HTTP surface

- unified endpoint `/proxy` уже существует и используется как HTTP fallback path для composable mounting
- alias `/_ws` уже монтируется app-specific binaries (`network-debugger-web`, `wsapp`) и считается binary/app compatibility surface, а не public default path
- legacy REST surface всё ещё жива и покрыта тестами:
  - `/api/version`
  - `/api/sessions`
  - `/api/sessions/{id}`
  - `/api/sessions/{id}/ws/send`
  - `/api/monitor/ws`
  - `/api/sessions_stream/...`

Вывод:

- extraction не должен случайно объявить эти пути частью public `proxykit` API
- но app adapters обязаны сохранить их до конца migration cycle

### Realtime monitor protocol contract

Frontend inspector сейчас уже ожидает не просто поток monitor events, а определённый delivery protocol:

- Socket.IO endpoint именно `/_api/v1/monitor/io/socket.io/`
- server sends initial snapshot even before explicit resubscribe
- filters payload включает:
  - `q`
  - `target`
  - `types`
  - `status`
  - `tags`
  - `captureScope`
  - `includeUnassigned`
  - `limit`
  - `groupBy`
- connection-specific aggregates считаются отдельно и эмитятся debounce-ом
- details UI использует room protocol:
  - `session:subscribe`
  - `session:unsubscribe`
  - `session:frames`
  - `session:events`
  - `session:http`
- при проблемах realtime frontend откатывается на REST loading

Вывод:

- это frozen app compatibility contract на период миграции
- public `observe` не должен фиксировать ни эти event names, ни room semantics, ни aggregate payload shape

### Dual monitor channels contract

- frontend одновременно использует:
  - Socket.IO channel for sessions/aggregate/detail updates
  - raw WS monitor channel for global events and intercept queue refresh
- raw monitor listeners уже ожидают:
  - `sessions_cleared`
  - `session_error`
  - `intercept_*`
- Socket.IO listeners при этом не заменяют raw monitor consumers

Вывод:

- current channel split - это app delivery architecture
- extraction не должен пытаться "красиво объединить" это в один public channel abstraction без adapter layer

### Compose and replay contract

Текущее приложение работает сразу с двумя разными способами "пустить HTTP запрос":

- inline replay/refetch из HTTP inspector идёт через `/httpproxy`
- compose идёт через `/_api/v1/compose/send` и создаёт synthetic session/frames/http transaction без использования reverse proxy path
- `_resetCapture=true` на первом reverse request уже тоже часть app workflow и используется helper packages как orchestration signal, а не transport metadata

Вывод:

- session model приложения нельзя сводить к "всё пришло через proxy handlers"
- app adapters обязаны сохранить оба сценария
- public proxy foundation не должен случайно втянуть compose DTO/history/library API в transport packages

### Capture and visibility contract

- backend state `recording/current capture` живёт в store, а не в proxy runtime
- listing/realtime semantics уже завязаны на:
  - `captureId=current`
  - `captures=all`
  - `includeUnassigned`
  - `captureScope`
  - `pausedSince`
- `POST /_api/v1/capture`, `GET /_api/v1/capture`, `GET /_api/v1/captures`, `POST /_api/v1/capture/reset` уже часть frontend contract
- reset capture должен делать не только state flip, но и:
  - clear sessions
  - close live connections
  - emit `sessions_cleared`
  - start new capture id

Вывод:

- capture lifecycle - это app-layer state machine and visibility policy
- public core может быть unaware of current-capture UX and reset behavior

### Ingest and admin-tooling contract

- `firebase_database_debugger` package уже использует `/_api/v1/ingest/firebase_database`
- ingest endpoint поддерживает explicit capture binding и отдельную auth policy
- `X-Admin-Token` уже shared contract для mapping/intercept/ingest and related frontend clients
- CORS allow-headers уже включают `X-Admin-Token`

Вывод:

- protocol-specific ingest и admin auth policy - это app edge, не core proxy concern

### Binary delivery and TLS contract

- `network-debugger` binary - backend/proxy with informational root `/`
- `network-debugger-web` и `wsapp` - embedded SPA binaries with static fallback on `/`
- optional TLS listener уже intentionally обслуживает REST/reverse отдельно от forward proxy/CONNECT
- `withCORS` живёт на уровне app router wrapping, а не внутри extracted transport engines

Вывод:

- binary delivery, SPA mounting, CORS wrapping и TLS listener split - это app composition layer
- public packages должны оставаться headless и route-agnostic

### WebSocket contract

- `/wsproxy` и `/wsproxy/` должны обе работать
- `_target=http(s)://...` должен автоматически нормализоваться в `ws(s)://...`
- `Authorization`, `Cookie`, `Origin`, `User-Agent`, `Referer`, `Sec-WebSocket-Protocol` должны корректно проходить до upstream
- binary frames должны проксироваться так же надёжно, как text frames
- один connection = одна реальная ws-session в storage/monitoring модели приложения

### Live WS control contract

- legacy/app route `/api/sessions/{id}/ws/send` уже существует и использует live registry, а не historical storage
- приложение умеет отправлять manual text frames по направлению:
  - `client->upstream`
  - `upstream->client`
- capture reset, clear-all и related cleanup flows уже зависят от закрытия active live WS sessions

Вывод:

- live session registry and manual-send surface - это app debugging/control contract
- extracted `wsproxy` transport не должен брать на себя ownership этого API или route shape

### Socket.IO compatibility

- Socket.IO mode реально использует тот же `/wsproxy`, а не отдельный transport
- namespace path и engine.io path semantics нельзя терять при extraction
- `SOCKET_UPSTREAM_TARGET`, `SOCKET_UPSTREAM_URL`, `SOCKET_UPSTREAM_PATH` уже участвуют в attach logic Dart package
- reverse attach не должен ломать namespace path, когда proxy host совпадает с upstream host и upstream берётся из env/define

### Cookie boundary contract

- per-request `_cookie_mode` override уже поддерживается поверх runtime config
- режимы `off/auto/isolate` уже часть observable behavior
- namespace изоляции считается от `scheme://host[:port]`, а не от full URL path/query
- mount prefix влияет на cookie `Path`, то есть `/httpproxy` и `/proxy` реально дают разную rewrite semantics
- для `__Host-` и `__Secure-` cookies должны сохраняться RFC-sensitive ограничения
- `SameSite=None` добавляет `Secure` только если client->proxy реально HTTPS
- `proxyHost` strategy не должна выставлять `Domain` для `localhost` и IP addresses
- неизвестные cookie attributes вроде `Priority`, `Partitioned`, кастомных токенов и `Expires` c запятой нельзя терять при rewrite
- в isolate mode outbound `Cookie` header должен отправлять upstream только cookies текущего namespace

Вывод:

- cookie policy уже нельзя считать мелкой utility-функцией
- при extraction её надо держать как отдельный boundary с explicit regression suite

### Session, httpMeta и body access contract

Frontend и тесты уже опираются не на один источник метаданных, а на комбинацию:

- `session.httpMeta` из REST/WebSocket payload
- lazy warmup `httpMeta` через `/_api/v1/sessions/{id}/frames`
- raw body fetch через `/_api/v1/sessions/{id}/frames/{frameId}/body`
- fallback поведения body endpoint:
  - если `BodyFile` отсутствует, вернуть preview
  - если файл есть, вернуть raw bytes
  - если файл удалён GC/cleanup, вернуть controlled error

Вывод:

- `httpMeta` и `sizes` это уже внешний UI contract
- serving raw body нельзя случайно потерять при выделении capture/observe слоёв
- app adapter обязан сохранить dual-source модель `HTTPTransaction -> frames fallback`
- sessions list UI использует eager `session.httpMeta`, но при этом детали и часть фильтров опираются на shared cache/warmup из frames
- значит adapter contract тут не одинарный, а dual-path: server-provided meta + lazy enrichment
- `httpMeta` уже несёт derived fields, а не только raw transport facts:
  - `cache`
  - `cors`
  - `preflight`
  - user-facing error fields

### Structured HTTP preview contract

- HTTP request/response frames уже содержат JSON preview с stable discriminator:
  - `type=http_request`
  - `type=http_response`
- preview shape уже используется UI-слоями для:
  - body truncation detection через `bodyRawSize`
  - request/response info tabs
  - copy-as-cURL and Compose conversion через `headersRaw`
  - cookie and TLS summaries
  - timings enrichment
  - `httpMeta` warmup fallback
- masking/raw-header duality (`headers` vs `headersRaw`) уже часть app privacy UX

Вывод:

- current preview JSON shape - это frozen app compatibility contract на период миграции
- public core должен давать raw observations, а не тащить этот exact preview schema в foundation

### Frame body endpoint contract

- `/_api/v1/sessions/{id}/frames/{frameId}/body` уже поддерживает несколько meaningful outcomes:
  - raw file body
  - preview fallback when body file absent
  - `404 FRAME_NOT_FOUND`
  - `410 BODY_FILE_EXPIRED`
  - `413 BODY_TOO_LARGE`
- response headers already matter:
  - `X-Body-Source=file|preview`
  - `X-Frame-Id`
- frontend body panels already branch on these semantics, not just on bytes presence

Вывод:

- body endpoint - это app compatibility surface, а не internal storage helper
- extraction body/capture layers должен сохранить status/header semantics этого endpoint

### Session V1 projection contract

- `/_api/v1/sessions` уже является projected query API, а не сырым storage list:
  - `types`
  - `status`
  - `tags`
  - `scan=graphql`
  - `captureId/current`
  - `captures=all`
  - `includeUnassigned`
- response items already include app-shaped projections:
  - `httpMeta`
  - `sizes`
  - `processInfo`
- frontend list/realtime/filter flows и `nd_test_support` already use this as the main sessions surface

Вывод:

- list/query projection semantics - это app layer
- public `observe` не должен стабилизировать current filter vocabulary or projected response shape

### Error projection contract

- `session_error` monitor payload и `httpMeta.error*` уже образуют совместный UI contract
- user-facing codes/messages сейчас derived from app-side classification, не из raw `error.Error()`
- cancellations intentionally treated differently from real errors in notifications and filters
- HTTP and WS errors уже расходятся по presentation policy

Вывод:

- error classification/suppression/presentation - это app concern
- public foundation не должен prematurely stabilise current `errorCode/errorCategory/userMessage` model as core API

### Proxy runtime config contract

UI и backend уже согласованы по конкретной модели runtime config:

- `/_api/v1/proxy/config`
- `forward.enabled/port/addr`
- `socks.enabled/port/addr/authMode`
- validation на диапазон портов
- validation на `authMode`
- запрет на одинаковый порт для forward и socks

Вывод:

- это frozen app contract на период миграции
- `proxykit` не должен тащить этот JSON contract в public API
- separation должна проходить так:
  - `proxykit/proxyruntime` - lifecycle и apply semantics
  - app adapter/API - DTO, validation, persistence, UX

### Proxy runtime credential/privacy contract

- GET `/_api/v1/proxy/config` intentionally omits SOCKS credentials
- POST behaves like app-level partial update for credentials rather than full replace
- settings/integrations UI already rely on config DTO for port discovery while keeping secrets out of normal read flow

Вывод:

- credential redaction and partial-update semantics are part of app API policy
- public runtime package should stay unaware of this privacy/UX contract

### Runtime simulation contract

- persisted runtime knobs уже являются частью settings UX:
  - response delay
  - throttle on/off
  - up/down bandwidth
  - packet loss
  - latency/jitter
  - offline mode
- `/_api/v1/throttle` и `/_api/v1/throttle/profiles` уже внешний app contract
- persisted throttle profiles уже имеют user-facing identity, а не просто transient config blobs
- одна и та же simulation policy должна одинаково влиять на reverse, forward и WS flows
- frontend settings page уже считает эти knobs persisted feature set, а не временным debug state

Вывод:

- runtime simulation - это frozen app contract на период миграции
- extraction не должен размазать эти semantics по нескольким public transport packages
- если позже выносить наружу, то отдельным optional package с explicit API, а не как incidental options on `reverse`/`forward`

### MITM dev-tooling contract

- `/_api/v1/mitm/status`, `/_api/v1/mitm/ca`, `/_api/v1/mitm/ca/generate`, `/_api/v1/mitm/ca/regenerate` уже часть app-layer tooling surface
- при `MITMEnabled` backend сейчас умеет автогенерировать и сохранять dev CA, если пути не заданы
- allow/deny suffix lists уже участвуют в runtime behavior
- текущая реализация по умолчанию хранит dev CA в `data/mitm_dev_ca.{crt,key}`, а не в `--data-dir`, значит file-location policy уже app-specific и её нельзя молча менять во время extraction

Вывод:

- MITM runtime можно выделять только как optional package
- CA lifecycle, HTTP endpoints, storage path policy и UX messaging должны оставаться в app layer

### Frontend UX compatibility

- integrations/settings UI считает `/_api/v1/proxy/config` источником истины для forward proxy port
- HTTP inspector уже делает inline refetch и replay через `/httpproxy`
- `nd_test_support` ожидает, что Go binary можно собрать как `./cmd/network-debugger`, запустить локально и опрашивать по health endpoint
- compose UI сейчас тоже опирается на `frames` API и минимальный `httpMeta` fallback, а не на отдельный compose-specific transport contract
- sessions UI и filters опираются на derived `httpMeta`/tags/status groups, а не только на raw session kind

### Process enrichment and helper contract

- settings UI уже ожидает stable surfaces:
  - `/_api/v1/process/config`
  - `/_api/v1/process/helper/status`
  - `/_api/v1/process/helper/install`
- `Session.ProcessInfo` уже часть observable session payload, а не internal debug detail only
- helper install/status/version, cache TTL and fallback behavior are already product semantics, not just implementation details

Вывод:

- process detection/config/install flow - это app integration contract
- extraction не должен протащить helper policy, platform install flow или local permission story в public proxy packages

### Tags and annotations metadata contract

- predefined tags, session tags and session annotations already form a stable app surface
- selected tags persist in prefs и участвуют в session filtering and realtime subscribe payloads
- frontend/Dart clients already expect fallback behavior between per-session tags API and bulk tags API

Вывод:

- tags/annotations contract - это app metadata/filtering layer
- public transport/observe foundation не должен объявлять его частью reusable proxy API

## Наименее надёжные зоны и что с ними делать

Ниже не просто риски, а зоны, где сейчас самая низкая уверенность, если делать extraction "по наитию".

### 1. Ordering pipeline в reverse HTTP

Сейчас в одном handler перемешаны:

- path/query assembly
- mapping
- scripting
- interception
- cookie rewriting
- preview generation
- transaction persistence
- spool body files
- error mapping

Это значит, что главный вопрос не "как вынести reverse", а "в каком порядке должны жить mutation stages".

Надёжное решение:

- сначала ввести явный pipeline contract
- отдельно описать `request stages` и `response stages`
- только потом выносить reusable reverse engine

### 2. Capture vs transport boundary

Сейчас реальные app scenarios завязаны и на:

- session list
- frame list
- events
- HTTP transaction summary
- preview snippets
- body file fetch

Надёжное решение:

- не пытаться сразу сделать один public `capture` пакет со всей app-моделью
- сначала сделать `observe contracts`
- затем `network-debugger adapters`
- только после этого думать, какая часть capture model реально универсальна для сообщества

### 3. Cookie isolation

Текущая cookie logic уже достаточно сложная, чтобы испортить public API одним неверным abstraction:

- isolate/auto/off
- namespace
- domain strategy
- path strategy
- secure/samesite нюансы

Надёжное решение:

- cookie policy не включать в `reverse` v1 extraction как жёсткую встроенную модель
- сначала оставить app adapter над текущей логикой
- потом выделить `cookiepolicy` пакет отдельно, когда будет понятна минимальная стабильная API surface

### 4. Forward + CONNECT + MITM + WS-over-forward

Это не один engine, а набор связанных, но разных responsibilities.

Надёжное решение:

- `forward` package отвечает за regular HTTP forward proxy и CONNECT tunnel
- WS-over-forward должен быть отдельной extension capability
- MITM должен жить отдельно от `forward`

### 5. Response body lifecycle

В reverse/forward коде уже есть тонкая логика:

- preview читает body
- capture/spool читает body
- intercept может менять body
- затем body надо вернуть клиенту живым

Надёжное решение:

- до extraction `reverse` и `forward` надо зафиксировать единый body-buffering contract
- нельзя оставлять body lifecycle как набор ad-hoc `ReadAll + restore`

### 6. `httpMeta` и body-serving seam

Сейчас UI и tests живут на смешанной модели:

- `HTTPTransaction` даёт method/status/mime/duration/sizes
- frame previews дают fallback, когда transaction ещё нет
- отдельный endpoint для raw body зависит от `Frame.BodyFile`

Надёжное решение:

- `observe` слой не должен тащить внутрь `httpMeta` как UI DTO
- app adapter должен отдельно собирать `httpMeta` view model
- raw body serving и file cleanup policy должны остаться в приложении до появления реально reusable body-store abstraction

### 7. Public API shape для Go users

Если делать extraction только по слоям, но без жёстких правил для export surface, модуль получится технически рабочим, но community-grade не будет.

Надёжное решение:

- public packages не читают env и не зависят от process globals
- prefer concrete exported types + `New(...)` / option structs
- не экспортировать producer-side широкие interfaces без необходимости
- API должен расширяться через новые поля config/option structs и новые hooks, а не через ломающие interface methods
- избегать generic package names вроде `core`, `common`, `util`, `helpers`

### 8. Mutation orchestration boundary

Самая низкая уверенность сейчас не в transport, а в orchestration порядка:

- mapping может переписать upstream и host
- scripts могут менять request/response
- intercept может поставить blocking decision queue
- cookies и capture зависят от уже изменённого request/response

Надёжное решение:

- в public `reverse` engine закладывать generic mutation phases
- но не закладывать туда app-shaped `mapping rule`, `script`, `intercept item`
- порядок фаз должен быть закреплён отдельным contract document внутри плана и будущих tests

### 9. Temp-file ownership и cleanup semantics

Сейчас один и тот же body/file path участвует в:

- session details body loading
- session deletion
- clear-all
- imported cleanup
- spool GC TTL
- HAR export include-bodies

Надёжное решение:

- не прятать file lifecycle внутрь public transport package
- app adapter обязан быть owner cleanup semantics
- если позже появится reusable abstraction, она должна моделировать ownership и retention явно

### 10. Startup composition boundary

Router/startup код сейчас отвечает сразу за:

- settings overlay над env-config
- lazy init сервисов
- proxy runtime apply
- mapping runtime preload
- intercept seed/load
- spool GC start

Надёжное решение:

- `proxykit` не должен получить "god bootstrap"
- startup composition и service wiring остаются app concern
- public packages должны быть composable pieces без знания full application boot order

### 11. Documentation and examples are part of the public contract

Для community-grade модуля качество оценивается не только кодом, но и тем, насколько package docs/examples соответствуют реальным сценариям.

Надёжное решение:

- каждый public package должен иметь minimal, runnable examples
- examples должны покрывать reverse, forward, WS and composable mounting story
- docs не должны копировать app-specific routing as the only usage model

## Архитектурные решения, которые фиксируем

### 1. Toolkit-first, не framework-first

Идём в:

- small reusable packages
- optional facade/adapters
- `network-debugger` как reference app

Это соответствует и требованиям пользователя, и Go ecosystem.

### 2. Сначала public backend identity, потом app convenience

Не допускаем в public module:

- `/_api/v1/...` DTO
- frontend monitor payloads
- `SessionService`
- GORM/SQLite storage assumptions
- Flutter/Dart naming
- app-specific errors и endpoint paths

### 3. MITM не входит в core

MITM полезен, но его надо держать отдельно:

- отдельный пакет
- отдельные tests
- отдельные API guarantees

Причина:

- большой security surface
- отдельные причины изменения
- высокая цена регрессий

### 4. Capture остаётся hooks-first

Core transport packages должны эмитить neutral observations/hooks, а не писать в storage сами.

Это обязательно, иначе мы снова получим повтор `SessionService`-centric архитектуры.

### 5. Nested module only for incubation

Сильное community-grade решение:

- использовать nested module только как переходный stage для extraction и regression safety
- не публиковать stable releases из Flutter-centric mono-repo
- до первого реального публичного релиза вынести `proxykit` в отдельный repo

Это лучше и по ergonomics для пользователей, и по release discipline.

### 6. Public API должен быть idiomatic Go, не "internal backend made public"

Фиксируем такие правила:

- package names должны читаться естественно в import-site
- core packages не знают про env parsing, CLI flags, HTTP routes и storage
- package docs и `Example` tests считаются частью API quality, а не nice-to-have
- exported surface должен быть маленьким и объяснимым без чтения исходников приложения

### 7. SOLID boundary по акторам, а не по protocol buzzwords

Разделение делаем по причинам изменения:

- transport actor
- observation/capture actor
- mutation orchestration actor
- runtime lifecycle actor
- runtime simulation actor
- app API/UI compatibility actor
- app projection/query actor
- startup composition actor
- monitor/aggregation actor
- import/export actor
- local machine integration actor
- privacy/security policy actor
- session classification/filtering actor
- session metadata/tags actor
- documentation/examples actor

Не делаем разделение по принципу:

- "всё HTTP в один package"
- "всё proxy в один package"
- "всё hooks в один package"

## Что остаётся в приложении

- REST API
- `/httpproxy`, `/wsproxy`, `/_api/v1/proxy/config`
- `/proxy`, `/_ws`, `/_api/v1/httpproxy`
- legacy `/api/*` routes и root `/` landing page
- monitor WebSocket и frontend events
- session/frame/event storage
- session list enrichment в app-shaped `httpMeta` и `sizes`
- structured HTTP preview JSON schema for inspector/compose UX
- raw body serving endpoint и cleanup policy для spool/body files
- body endpoint status/header semantics like `X-Body-Source`, `410`, `413`
- proxy runtime DTO/validation/persistence for settings UI
- proxy runtime credential redaction and partial-update policy
- browser auto-open policy, launcher env contract and readiness UX
- realtime Socket.IO delivery protocol, rooms and REST fallback policy
- binary-specific route surfaces, SPA mounting and CORS wrapper policy
- compose synthetic session generation and history/library API
- capture lifecycle, current-capture visibility model and reset workflow
- protocol-specific ingest endpoints and admin-token auth policy
- dual monitor channels and event delivery topology
- user-facing error classification, suppression and notification policy
- response delay, throttle profiles, offline/latency/packet-loss runtime simulation UX
- live WS registry and manual WS send/control surface
- HTTP transaction persistence
- DB-backed runtime settings
- process detection, helper install/status/config and session process enrichment
- startup composition and boot order
- compose
- compose session details projection over frames/httpMeta
- mapping persistence
- mapping orchestration and rule management
- scripting persistence and execution wiring
- scripting runtime registration/validation/plugin lifecycle
- intercept queue/timeout/overflow workflow
- intercept persistence/config UI
- monitor hub, Socket.IO subscription protocol, aggregate updates and session rooms
- HAR import/export and imported session semantics
- projected `/_api/v1/sessions` query/filter semantics and response shape
- session classification, quick filters, GraphQL deep-scan and derived tags
- predefined tags, session tags, session annotations and tags-based filtering/prefs
- performance analytics built over app-shaped session projections

## Что должно стать reusable

### Уже reusable

- listener runtime
- Socket.IO parser
- shared HTTP proxy utilities
- reusable WS proxy engine

### До extraction HTTP engines нужен отдельный contracts layer

Новый обязательный промежуточный слой:

- `proxykit/observe`

Или, если имя будет точнее после реализации:

- `proxykit/contracts`
- `proxykit/events`

Но смысл один:

- session lifecycle contracts
- request/response observation
- WS frame observation
- optional protocol event observation

Без этого `reverse` и `forward` почти гарантированно потащат app-shaped зависимости.

### Следующие кандидаты на extraction

#### Phase 2 - `observe` contracts

Должен взять на себя:

- neutral lifecycle/event contracts
- extensible option structs
- transport-safe callback interfaces
- backward-compatible expansion strategy через новые struct fields и новые hooks, а не через breaking function signatures
- neutral request/response snapshots для observation
- body references/handles только как transport-facing abstraction, не как storage schema
- explicit distinction between:
  - bounded inline payload preview
  - body reference/handle
  - transport metadata
- sufficient raw inputs so app can derive:
  - cache meta
  - CORS/preflight meta
  - session tags/classification
  but without baking those derived concepts into core

Не должен знать:

- SQLite
- GORM
- frontend DTO
- app session storage
- `httpMeta` view model
- current HTTP preview JSON schema
- `frames/{id}/body` endpoint semantics
- конкретные file paths и cleanup policy приложения
- app-specific tags/status-group/filter model

#### Phase 3 - `reverse`

Должен взять на себя:

- upstream target resolution abstraction
- request path/query assembly
- reverse transport execution
- request/response mutation phases
- shared error handling contracts
- composition-friendly API so app can mount reverse handler on multiplexed proxy listener

Не должен знать:

- mapping storage
- script storage
- spool files
- UI endpoints
- cookie persistence
- app monitor payload shapes
- intercept queue workflow
- script executor lifecycle
- monitor room protocol
- HAR import/export semantics

#### Phase 4 - `forward`

Новый public package:

- `proxykit/forward`

Должен взять на себя:

- absolute URI request handling
- CONNECT tunneling
- request/response forwarding
- optional WS-over-forward bridge

Не должен знать:

- app monitor contracts
- frontend API paths
- MITM certificate management

#### Phase 5 - optional packages

- `proxykit/cookies` или `proxykit/cookiepolicy`
- `proxykit/observe` или `proxykit/capture`
- `proxykit/mitm`
- `proxykit/netsim` или другой отдельно названный simulation package для latency/bandwidth/loss/offline policy

Но только после extraction raw engines и observer contracts.

## Следующий порядок реализации

### Этап A - стабилизировать новый public module

- держать `proxykit` с package docs, tests, README
- не раздувать API
- все новые extraction-и делать сразу туда, а не в `internal`
- проверить package naming до появления новых exports, чтобы не тащить наружу слабые имена
- не допускать env-driven configuration в public core packages
- examples должны показывать configurable mounting and composition, not only current app defaults

### Этап B - выделить `wsproxy`

Сделано.

Почему это был правильный первый transport extraction:

- меньше зависимостей, чем у HTTP handlers
- у него уже есть естественный observer seam
- он критичен для community value
- он нужен и текущему приложению, и внешним Go users

### Этап C - построить neutral observer contracts

Нужны минимальные контракты для:

- session open
- session error
- session close
- WS frame observation
- optional protocol event observation
- HTTP request/response observation
- extensible option structs для не-breaking growth

Правило:

- contracts должны быть transport-neutral или почти neutral
- никаких `domain.Frame` и `domain.Event` из текущего приложения в public API
- новые возможности добавляются через новые поля в option/config structs, а не через ломающие изменения сигнатур
- observation payloads должны быть достаточными для app adapters, но не обязаны совпадать с текущими frontend DTO
- body access в hooks должен быть либо bounded buffer, либо explicit body handle, но не скрытая storage модель
- надо явно различить session summary, frame/event observation и body reference
- `httpMeta` остаётся responsibility app adapter, а не observer layer

Ориентир:

- [Keeping Your Modules Compatible](https://go.dev/blog/module-compatibility)

### Этап D - перевести `network-debugger` на app adapters

Создать app-layer adapters, которые мапят public hooks в:

- `SessionService`
- `MonitorHub`
- metrics
- live session registry
- existing DTO/events
- `httpMeta` и `sizes` enrichment
- `frames/{id}/body` serving поверх текущего body spool policy
- proxy settings DTO + runtime apply flow
- mapping/intercept/script orchestration поверх generic mutation/observe seams
- multiplexed proxy-port routing
- monitor hub / aggregate updates / session room protocol
- HAR import/export and imported-session workflows
- session classification/tags/status-group filters
- cache/CORS/preflight/error projections inside app-shaped `httpMeta`

### Этап E - только потом резать HTTP handlers

HTTP extraction делать после того, как observer contracts и WS path стабилизированы.

Причина:

- reverse и forward намного плотнее переплетены с mapping/script/intercept/cookies
- без observer contracts получится плохой public API

### Этап F - перед публичным release вынести `proxykit` в отдельный repo

Это не optional nice-to-have, а целевой publish gate.

Перед первым внешним release нужно:

1. вынести `proxykit` в отдельный repo
2. сохранить тот же import path
3. перевести root repo на внешнюю зависимость или `go work` во время совместной разработки
4. завести отдельный release workflow и examples именно для `proxykit`

## Риски и способы удержания качества

### Риск 1 - вынести слишком много unstable API

Контрмера:

- export только то, что уже нужно приложению и внешним users
- internal helper-ы не торчат наружу

### Риск 2 - повторно зацементировать app-shaped capture model

Контрмера:

- hooks вместо storage contracts
- adapters в приложении, не в модуле

### Риск 3 - размыть package responsibilities

Контрмера:

- отдельная причина изменения на пакет
- MITM, cookies, observe не смешивать с raw transport

### Риск 4 - сломать frontend и Dart SDK packages

Контрмера:

- compatibility endpoints живут до конца миграции
- frontend adaptation идёт в том же цикле, что и backend changes
- health endpoints `/healthz` и `/_health` считаются frozen compatibility contract
- CLI flags для локального launcher flow тоже считаются frozen contract на период миграции
- route aliases `/proxy`, `/_ws`, `/_api/v1/httpproxy`, `/readyz` и legacy `/api/*` тоже нельзя терять без adapter migration

### Риск 5 - потерять regression coverage

Контрмера:

- переносить и дублировать критичные tests в `proxykit`
- старые internal tests временно сохранять через shims

### Риск 6 - выпустить неудобный public module из nested repo

Контрмера:

- nested module только для staging
- отдельный repo до публичного release
- не ставить stable tags на submodule внутри app-centric repo

### Риск 7 - сломать edge-case semantics, которые уже зафиксированы в Dart tests

Контрмера:

- explicit compatibility checklist по query merge, trailing slash, auto-normalization, repeated query keys, redirect rewrite, header forwarding, no double-proxying

### Риск 8 - сделать формально reusable модуль, но с не-идиоматичным Go API

Контрмера:

- package naming review до добавления новых exports
- examples и package docs вместе с каждым новым public package
- маленькие concrete APIs вместо расплывчатых umbrella interfaces

### Риск 9 - сломать session/httpMeta/body UX без прямой поломки transport

Контрмера:

- explicit tests на `httpMeta` enrichment
- explicit tests на `frames/{frameId}/body` fallback/file/expired scenarios
- app adapters не выносить в `proxykit`

### Риск 10 - случайно встроить queue-oriented intercept workflow в public transport API

Контрмера:

- `reverse` и `forward` знают только про generic mutation phases
- pending queue, timeout, overflow, continue/cancel остаются app concern до отдельного зрелого extraction

### Риск 11 - смешать body observation с temp-file ownership

Контрмера:

- `observe` говорит про payload/reference, а не про temp files
- cleanup и retention policy остаются в app adapter

### Риск 12 - сломать settings/proxy runtime UX без поломки listener runtime

Контрмера:

- отдельные tests на `/_api/v1/proxy/config`
- DTO/validation contract фиксируем как app-layer concern
- `proxyruntime` остаётся endpoint-agnostic

### Риск 13 - потерять multiplexed proxy-port behavior

Контрмера:

- отдельно фиксируем, что proxy port в приложении остаётся composition of:
  - reverse HTTP
  - reverse WS
  - forward HTTP/CONNECT
  - health endpoints

### Риск 14 - сломать live monitor protocol без поломки transport

Контрмера:

- Socket.IO subscription/aggregate/session-room semantics остаются app concern
- tests нужны не только на transport events, но и на adapter projection в monitor protocol

### Риск 15 - сломать HAR/imported-session сценарии при рефакторинге body ownership

Контрмера:

- imported sessions и HAR body flows считаются отдельным compatibility gate
- body ownership changes нельзя делать без explicit regression tests на import/export

### Риск 16 - случайно зацементировать derived UI metadata в public observe API

Контрмера:

- `httpMeta`, tags, status groups, GraphQL flags, cache/CORS/preflight projections остаются app concern
- public observe model даёт raw enough data, but not UI semantics

### Риск 17 - документация public module будет слабее, чем у существующих Dart helpers

Контрмера:

- docs/examples считаются quality gate before public release
- examples должны покрывать path-agnostic mounting and common reverse/forward/WS stories
- release не делаем, пока public docs хуже по ясности, чем текущие helper package docs

### Риск 18 - случайно сломать startup/browser semantics, не затронув transport

Контрмера:

- держать browser auto-open и launcher env contract в app layer
- отдельно тестировать `/_health`, `/readyz`, root `/` landing page и desktop launcher expectations
- не переносить `NO_BROWSER`/`OPEN_BROWSER`/`DEV_MODE` semantics в `proxykit`

### Риск 19 - упростить cookie layer так, что сломаются prefix/path/namespace edge cases

Контрмера:

- explicit tests на `/_cookie_mode`, `/httpproxy` vs `/proxy`, `__Host-`, `__Secure-`, `SameSite=None`, localhost/IP domain fallback, unknown attribute preservation
- cookie policy extraction делать только после фиксации minimal public contract

### Риск 20 - упростить MITM extraction и потерять dev-tooling lifecycle

Контрмера:

- MITM package держать optional и headless
- status/generate/regenerate endpoints и CA file policy оставлять в app layer
- изменения вокруг CA location и regen flow делать только с явной migration strategy

### Риск 21 - случайно зацементировать current realtime UI protocol в public observe API

Контрмера:

- отделять raw observations от delivery protocol
- Socket.IO event names, room lifecycle, aggregate debounce и REST fallback держать в app adapters
- tests на realtime protocol писать в app layer, не как public package contract

### Риск 22 - предположить, что все sessions происходят только из proxy transport

Контрмера:

- учитывать synthetic producers вроде compose отдельно
- не тащить compose DTO/history/library в public proxy packages
- observe contracts проектировать так, чтобы app мог объединять proxy-produced и synthetic sessions без загрязнения core API

### Риск 23 - потерять binary-specific route surface или TLS split при "упрощении" server composition

Контрмера:

- отдельно тестировать `network-debugger`, `network-debugger-web`, `wsapp` route surfaces
- сохранять boundary `plain forward/CONNECT listener` vs `TLS reverse/REST listener`
- CORS/SPA mounting держать в app composition, не в transport engines

### Риск 24 - зацементировать current capture/paused visibility model в public observe API

Контрмера:

- current/all/unassigned/paused semantics держать в app adapters and UI-facing query layer
- transport/observe core не должен знать про `captureScope`, `includeUnassigned`, `pausedSince`, `_resetCapture`
- tests на visibility model держать в app layer

### Риск 25 - смешать protocol-specific ingest with reusable proxy foundation

Контрмера:

- firebase ingest, HAR import and compose держать как synthetic producers поверх app session model
- не тянуть их DTO/auth/limits into proxykit core
- выделять отдельные adapters only if there is real community reuse signal

### Риск 26 - протащить AdminToken and loopback auth policy в public packages

Контрмера:

- `X-Admin-Token`, loopback/private bypass, endpoint auth policy оставить в app edge
- public packages не знают про app admin tooling security
- adapter tests нужны на protected endpoints and header propagation

### Риск 27 - объединить raw monitor WS и Socket.IO protocol в один public abstraction

Контрмера:

- current channel split держать в app adapters
- public `observe` описывает observations, не channel topology
- compatibility tests нужны отдельно на raw monitor consumers and Socket.IO consumers

### Риск 28 - зацементировать текущую UI error taxonomy в public core слишком рано

Контрмера:

- raw errors and low-level context can flow from core
- classification to `errorCode/errorCategory/userMessage` stays app-side until a truly reusable model emerges
- tests on notifications/filters/badges remain app-layer tests

### Риск 29 - размазать runtime simulation semantics по transport packages и потерять единое поведение

Контрмера:

- response delay, throttle, packet loss, latency/jitter and offline semantics держать как отдельный app contract
- tests нужны на одинаковое применение policy к reverse, forward и WS flows
- profiles/persistence/UI semantics нельзя смешивать с public transport options

### Риск 30 - случайно сделать live WS registry частью public `wsproxy` API

Контрмера:

- manual send, live registry and close-all flows держать в app adapters
- public `wsproxy` ограничивать transport hooks and connection lifecycle only
- tests на `/api/sessions/{id}/ws/send` и reset/close-live behavior держать в app layer

### Риск 31 - протащить process helper lifecycle в reusable proxy foundation

Контрмера:

- process enrichment contract отделять от transport observation
- helper install/status/config и platform-specific detectors держать в app layer
- tests на `ProcessInfo` projection and helper settings оставлять app-side

### Риск 32 - утащить tags/annotations/filter prefs в public observe model

Контрмера:

- tags/annotations считать app metadata over sessions
- public observe model оставлять raw enough for adapters, but not opinionated about labels/annotations/filter prefs
- compatibility tests на tags bulk fallback, annotations CRUD и selected-tags filters держать в app layer

### Риск 33 - зацементировать current HTTP preview JSON schema в public core

Контрмера:

- distinguish raw observations from app preview projection
- `http_request/http_response` preview JSON, `headersRaw`, `cookieSummary`, `tls`, `bodyRawSize`, `timings` держать в app adapters
- tests на preview schema and UI parsing оставлять app-side

### Риск 34 - сломать `frames/{id}/body` UX, думая что это просто file read

Контрмера:

- body endpoint semantics тестировать как compatibility contract
- `X-Body-Source`, `404`, `410`, `413` и preview fallback держать в app layer
- reusable body/capture abstractions не должны молча менять этот behavior

### Риск 35 - утащить projected `/_api/v1/sessions` query semantics в public observe/filter API

Контрмера:

- types/status/tags/graphql-scan/capture filters считать app query language
- public foundation держать raw and composable, without app-specific filter vocabulary
- tests на `/_api/v1/sessions` projection and filter behavior оставлять в app compatibility suite

### Риск 36 - смешать runtime apply с credential redaction/update policy

Контрмера:

- `proxyruntime` knows only listener config and apply lifecycle
- GET redaction, POST partial-update semantics и settings UX остаются в app API layer
- tests на proxy config secrecy and update behavior держать рядом с app endpoints

## Тестовая стратегия

### Уже проверено

- `go test ./...` в `proxykit`
- `go test ./internal/adapters/decoders/socketio ./internal/infrastructure/proxyruntime`
- targeted `httpapi` tests на helper/wiring contracts
- targeted `internal/integration` tests на WS scenarios после перевода `/wsproxy` на public engine

### Что должно стать обязательным compatibility suite

Нужно поддерживать отдельный список обязательных regression suites:

1. Dart e2e для `/httpproxy`
2. Dart e2e для `/wsproxy`
3. Dart e2e для Socket.IO over `/wsproxy`
4. backend integration tests на CONNECT
5. backend tests на query merge, redirect rewrite, header forwarding
6. frontend smoke checks на startup, proxy config UI, inline replay через `/httpproxy`
7. tests на `httpMeta` enrichment и `sizes`
8. tests на `frames/{frameId}/body` fallback/file/expired semantics
9. Dart package contract tests для `skip/allow` bypass semantics и no double-proxying
10. tests на `/_api/v1/proxy/config` DTO/validation/apply semantics
11. tests на spool cleanup via DeleteSession/ClearAll/GC-safe body endpoint behavior
12. tests на mapping chain semantics: remote rewrite, `StopProcessing`, `PreserveHost`
13. tests на proxy-port multiplexing composition
14. tests на monitor adapter protocol: `sessions:init/upsert/remove`, aggregate updates, session room events
15. tests на HAR import/export with imported sessions and body files
16. tests на derived app metadata: status groups, tags, GraphQL deep scan, cache/CORS/preflight projection
17. docs/examples review for public package usage clarity
18. tests на route aliases `/proxy`, `/_ws`, `/_api/v1/httpproxy`, `/readyz` и legacy `/api/*` surface
19. tests на cookie edge cases: prefix/path rewrite, namespace isolation, unknown attrs, localhost/IP domain fallback
20. startup/launcher smoke checks на `/_health`, root `/`, `NO_BROWSER`-driven flow и proxy-port health discrimination
21. tests на realtime protocol: `sessions:*`, `aggregate:update`, `session:*` rooms, reconnect fallback
22. tests на compose synthetic session flow and inspector replay/refetch coexistence
23. tests на binary-specific route surfaces and TLS split behavior
24. tests на capture lifecycle: `current/all/includeUnassigned`, `pausedSince`, `capture/reset`, `_resetCapture`
25. tests на firebase ingest/session visibility/auth behavior and admin-token protected APIs
26. tests на dual monitor channels: raw `monitor/ws` consumers vs Socket.IO consumers
27. tests на error projection: `session_error`, `httpMeta.error*`, cancellation suppression, WS-vs-HTTP handling
28. tests на runtime simulation semantics: response delay, throttle, packet loss, latency/jitter, offline and profile persistence across reverse/forward/WS
29. tests на live WS control: `/api/sessions/{id}/ws/send`, live registry cleanup and capture-reset close-all behavior
30. tests на process helper contract: config/status/install flows and `ProcessInfo` enrichment on created sessions
31. tests на tags/annotations contract: predefined tags, bulk fallback, annotations CRUD, selected-tags filtering and realtime subscribe payload
32. tests на structured HTTP preview contract: `http_request/http_response`, `headersRaw`, `bodyRawSize`, `cookieSummary`, `tls`, `timings`
33. tests на `frames/{id}/body` contract: `X-Body-Source`, preview fallback, `410 BODY_FILE_EXPIRED`, `413 BODY_TOO_LARGE`
34. tests на `/_api/v1/sessions` projection contract: filters `types/status/tags/scan`, `httpMeta`, `sizes`, `processInfo`
35. tests на proxy config privacy/update contract: GET redaction of credentials and POST partial-update behavior

### Что обязательно проверять на каждом следующем этапе

1. unit tests внутри `proxykit`
2. targeted adapter tests внутри `internal/infrastructure/httpapi`
3. integration tests на реальные proxy сценарии
4. Dart/Flutter contract tests после смены adapter wiring
5. `go test -race` хотя бы на `proxykit` и на targeted adapter packages
6. package examples должны компилироваться и отражать реальный public API
7. adapter tests на mutation phase ordering

## Quality Gates

Считаем extraction хорошим только если одновременно выполнено:

1. public package можно использовать без подтягивания Flutter/app-specific слоёв
2. текущий frontend не ломается
3. package names и imports выглядят по-Go-шному, а не как перенос UI backend вовне
4. transport packages не знают о storage
5. tests находятся рядом с новым public behavior, а не только в legacy app package
6. frozen compatibility contract по `/httpproxy`, `/wsproxy`, `/_api/v1/proxy/config`, `/healthz`, `/_health`, CLI flags не нарушен
7. до публичного release module publishing model приведена к single-repo-per-module
8. public API не требует знания `network-debugger` внутренних сущностей
9. package docs/examples достаточны, чтобы пользователь понял use-case без чтения mono-repo
10. mutation orchestration не зацементирована app-specific rule models в public packages
11. temp-file lifecycle не торчит в public transport API
12. startup composition, monitor protocol и HAR/import semantics остались в app layer
13. derived UI metadata/classification не зацементированы в public observe contracts
14. public docs/examples достаточно сильные, чтобы сообщество поняло reusable value без знания Flutter app
15. startup/browser env semantics и route alias surface остались в app layer и не протекли в public packages
16. realtime delivery protocol, compose synthetic flows и binary-specific server composition не зацементированы в public core
17. capture visibility model, synthetic ingest flows и admin-token security policy остались app concerns
18. dual monitor channel topology и current UI error taxonomy не зацементированы в public foundation
19. runtime simulation, live WS control, process helper lifecycle и tags/annotations metadata остались app concerns, а не утекли в public transport API
20. current HTTP preview schema, body endpoint semantics, session projection protocol и proxy-config privacy policy остались app contracts, а не протекли в public foundation

## Итоговый ориентир

Финальное состояние должно быть таким:

- `proxykit` - самостоятельный backend foundation для Go-сообщества
- `network-debugger` - приложение, которое использует `proxykit` через adapters
- текущие frontend contracts сохранены как compatibility layer
- новые protocol features добавляются расширением пакетов, а не ростом `internal/infrastructure/httpapi`
