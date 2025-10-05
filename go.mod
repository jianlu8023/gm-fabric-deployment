module github.com/jianlu8023/golang-example

go 1.22.0

toolchain go1.22.10

replace (
	github.com/Jeffail/gabs/v2 => github.com/Jeffail/gabs/v2 v2.7.0
	github.com/docker/docker => github.com/docker/docker v28.0.1+incompatible
	github.com/docker/go-connections => github.com/docker/go-connections v0.5.0 // 原本是0.4.0
	github.com/gin-contrib/cors => github.com/gin-contrib/cors v1.7.3
	github.com/gin-contrib/gzip => github.com/gin-contrib/gzip v1.2.2
	github.com/gin-contrib/requestid => github.com/gin-contrib/requestid v1.0.4
	github.com/gin-contrib/sse => github.com/gin-contrib/sse v1.0.0
	github.com/go-logr/logr => github.com/go-logr/logr v1.4.3
	github.com/go-logr/zapr => github.com/go-logr/zapr v1.3.0
	github.com/go-ozzo/ozzo-validation/v4 => github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-playground/validator/v10 => github.com/go-playground/validator/v10 v10.24.0
	github.com/golang-jwt/jwt/v5 => github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.5.3
	// github.com/ipfs/kubo => github.com/ipfs/kubo v0.28.0
	github.com/jianlu8023/go-logger/v2 => github.com/jianlu8023/go-logger/v2 v2.0.2
	github.com/jianlu8023/go-tools/v2 => github.com/jianlu8023/go-tools/v2 v2.0.0-20250930144632-d6dbaeab0b42
	github.com/juju/ratelimit => github.com/juju/ratelimit v1.0.2
	github.com/mingrammer/commonregex => github.com/mingrammer/commonregex v1.0.1
	github.com/mitchellh/mapstructure v1.4.2 => github.com/mitchellh/mapstructure v1.5.0
	github.com/mojocn/base64Captcha => github.com/mojocn/base64Captcha v1.3.8
	// github.com/pion/webrtc/v4 v4.0.9 => github.com/pion/webrtc/v4 v4.1.1
	github.com/scylladb/termtables => github.com/scylladb/termtables v0.0.0-20191203121021-c4c0b6d42ff4
	github.com/skip2/go-qrcode => github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/ulule/limiter/v3 => github.com/ulule/limiter/v3 v3.11.2
	github.com/unrolled/secure => github.com/unrolled/secure v1.17.0
	github.com/valyala/fasttemplate => github.com/valyala/fasttemplate v1.2.2
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc => go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.11.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp => go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.11.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc => go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp => go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.35.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog => go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.11.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric => go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.35.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace => go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.35.0
	go.opentelemetry.io/otel/exporters/zipkin => go.opentelemetry.io/otel/exporters/zipkin v1.35.0
	go.opentelemetry.io/otel/log => go.opentelemetry.io/otel/log v0.11.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.35.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/sdk/log => go.opentelemetry.io/otel/sdk/log v0.11.0
	go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.35.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.35.0
	golang.org/x/crypto => golang.org/x/crypto v0.33.0
	golang.org/x/net => golang.org/x/net v0.35.0
	golang.org/x/time => golang.org/x/time v0.10.0
	google.golang.org/genproto/googleapis/api => google.golang.org/genproto/googleapis/api v0.0.0-20241007155032-5fefd90f89a9
	gorm.io/driver/mysql => gorm.io/driver/mysql v1.5.7 // 原本是v1.5.1
	gorm.io/driver/postgres => gorm.io/driver/postgres v1.5.11 // 原本是v1.4.5
	gorm.io/gorm => gorm.io/gorm v1.30.0
	gorm.io/plugin/opentelemetry => gorm.io/plugin/opentelemetry v0.1.16
)

replace (
	// github.com/go-kit/kit => github.com/go-kit/kit v0.8.0
	// github.com/hxx258456/ccgo => github.com/hxx258456/ccgo v0.0.4
	// github.com/mitchellh/mapstructure => github.com/mitchellh/mapstructure v1.3.3
	// github.com/spf13/viper => github.com/spf13/viper v0.0.0-20150908122457-1967d93db724
	// go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20181228115726-23731bf9ba55
	// github.com/zmap/zcrypto => github.com/zmap/zcrypto v0.0.0-20190729165852-9051775e6a2e
	// github.com/zmap/zlint => github.com/zmap/zlint v0.0.0-20190806154020-fd021b4cfbeb
	// google.golang.org/grpc v1.64.0
	// google.golang.org/grpc v1.44.0
	google.golang.org/grpc => google.golang.org/grpc v1.64.0
	// google.golang.org/grpc => google.golang.org/grpc v1.63.2
	// google.golang.org/protobuf v1.28.1
	// google.golang.org/protobuf v1.34.2
	google.golang.org/protobuf => google.golang.org/protobuf v1.34.2
)

require (
	github.com/Jeffail/gabs/v2 v2.7.0
	github.com/casbin/casbin/v2 v2.122.0 // 权限控制
	github.com/docker/docker v27.3.0+incompatible
	github.com/docker/go-connections v0.5.0 // create docker container 需要这个库
	github.com/gin-contrib/cors v1.7.3
	github.com/gin-contrib/gzip v1.2.2
	github.com/gin-contrib/requestid v1.0.4
	github.com/gin-contrib/sse v1.0.0
	github.com/gin-gonic/gin v1.10.1
	github.com/glebarez/sqlite v1.11.0
	github.com/go-logr/logr v1.4.3
	github.com/go-logr/zapr v1.3.0 // logr 使用zapr
	github.com/go-playground/validator/v10 v10.25.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gorilla/websocket v1.5.3
	// github.com/hxx258456/ccgo v0.0.3 // grpc v1.44.0 protoc-gen-go-grpc 版本v1.2.0 没有grpc.NewClient 需要使用 grpc.Dial
	// github.com/hxx258456/fabric-sdk-go-gm v0.0.7
	// github.com/hyperledger/fabric-sdk-go v1.0.0
	// gitee.com/zhaochuninhefei/gmgo v0.1.1 // grpc升级到v1.63.2 protoc-gen-go-grpc 应该是v1.3.0
	// github.com/ipfs/boxo v0.27.2
	github.com/ipfs/go-cid v0.5.0
	github.com/ipfs/go-ipfs-api v0.7.0
	// github.com/ipfs/kubo v0.28.0
	github.com/jessevdk/go-flags v1.6.1 // flags增强 `short:"-v" long:"--version"  required:"true" default:"default"`
	github.com/jianlu8023/go-logger/v2 v2.0.0
	github.com/jianlu8023/go-tools/v2 v2.0.0-20250921142734-6afed502d952
	github.com/jinzhu/copier v0.4.0 // copy的功能 结构体 等值拷贝
	github.com/juju/ratelimit v1.0.2 // juju 限流
	github.com/libp2p/go-libp2p v0.40.0
	github.com/libp2p/go-libp2p-kad-dht v0.29.0
	github.com/mingrammer/commonregex v1.0.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mojocn/base64Captcha v1.3.8
	github.com/multiformats/go-multiaddr v0.14.0
	github.com/panjf2000/ants/v2 v2.11.3 // ants 异步方式线程池
	github.com/pion/interceptor v0.1.37
	github.com/pion/logging v0.2.3
	github.com/pion/webrtc/v4 v4.0.9
	github.com/pquerna/otp v1.5.0
	github.com/scylladb/termtables v0.0.0-20191203121021-c4c0b6d42ff4
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/sony/sonyflake v1.1.0
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.10.0
	github.com/tjfoc/gmsm v1.4.1
	github.com/ulule/limiter/v3 v3.11.2 // ulule 限流
	github.com/unrolled/secure v0.0.0-00010101000000-000000000000
	github.com/valyala/bytebufferpool v1.0.0
	github.com/valyala/fasttemplate v1.2.2
	github.com/valyala/quicktemplate v1.8.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.60.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.11.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.11.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.31.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.11.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.35.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.32.0
	go.opentelemetry.io/otel/exporters/zipkin v1.31.0
	go.opentelemetry.io/otel/log v0.11.0
	go.opentelemetry.io/otel/metric v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/sdk/log v0.11.0
	go.opentelemetry.io/otel/sdk/metric v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.35.0
	golang.org/x/time v0.5.0 // time的限流
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.5
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/postgres v1.5.11
	// github.com/Jeffail/tunny v0.1.4 // tunny 同步方式线程池
	// github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	// github.com/gin-contrib/cache v1.3.1
	// github.com/gin-contrib/sessions v1.0.2
	// github.com/pquerna/otp v1.5.0 // totp 验证码
	// github.com/gin-contrib/secure v1.1.1
	// github.com/gin-contrib/location v1.0.2
	// github.com/gin-contrib/authz v1.0.3
	// github.com/gin-contrib/i18n v1.2.2
	// github.com/gin-contrib/timeout v1.0.2
	// github.com/rsms/gotalk v1.3.7
	// github.com/emmansun/gmsm v0.29.8 // 解密x509证书 v0.27.3
	// github.com/sgoby/opencc v0.0.0-20181105060730-5b3b1de2620a // 翻译
	// gorm.io/driver/sqlite v1.6.0
	// gorm.io/driver/gaussdb v0.1.0 // toolchain go1.23.4 github.com/HuaweiCloudDeveloper/gaussdb-go
	// gorm.io/driver/sqlserver v1.6.1
	// gorm.io/driver/clickhouse v0.7.0
	// github.com/facebookgo/atomicfile v0.0.0-20151019160806-2de1f203e7d5 // 原子创建文件
	// github.com/RoaringBitmap/roaring/v2 v2.10.0 // bitmap
	// github.com/gorilla/schema v1.4.1 // schema 表单处理
	// github.com/gorilla/securecookie v1.1.2 // 加密cookie
	// github.com/gorilla/sessions v1.3.0 // session
	// github.com/markbates/goth v1.81.0 // 第三方认证
	// github.com/charmbracelet/bubbletea v1.3.4 // 控制台输出  spinner包 只显示文字
	// github.com/reactivex/rxgo/v2 v2.5.0
	// github.com/gin-contrib/pprof v1.4.0
	// github.com/unrolled/secure v1.14.0
	// github.com/PuerkitoBio/goquery v1.9.3 //类似于jquery
	// github.com/bamzi/jobrunner v1.0.0 // 运行job
	// github.com/robfig/cron/v3 v3.0.1 // cron 定时运行
	// github.com/urfave/negroni/v3 v3.1.1 // http中间件
	// github.com/smallnest/rpcx v1.8.32 // rpc库 toolchain go1.22.1
	// net/rpc/jsonrpc // jsonrpc 1.0 标准库
	// github.com/nutsdb/nutsdb v1.0.4 // 单机数据库
	// github.com/rs/zerolog v1.34.0 // 日志库
	// gonum.org/v1/plot v0.15.2 // plot 画图库
	// gopkg.in/h2non/gentleman.v2 // http库 https://github.com/h2non/gentleman
	// https://github.com/Knetic/govaluate // eval功能
	// github.com/jinzhu/copier v0.4.0 // copy的功能 结构体 等值拷贝
	// github.com/dave/jennifer v1.7.1 // 代码生成 链式调用
	// github.com/google/go-cmp v0.7.0 // 比较方法
	// github.com/golang-module/carbon/v2 v2.5.9 // 时间格式化
	// github.com/jordan-wright/email v4.0.0
	// github.com/vardius/message-bus v1.1.5 // 消息通信
	// github.com/ThreeDotsLabs/watermill v1.4.7 // 异步消息通信
	// github.com/joho/godotenv v1.5.1 // 读取.env 文件
	// github.com/spf13/cast v1.10.0 // cast方法
	// github.com/fsnotify/fsnotify v1.9.0 // 监听文件变化
	// github.com/mitchellh/go-homedir v1.1.0 // 获取home目录 为什么不适用os/user(需要cgo交叉编译)
	// github.com/gocolly/colly/v2 v2.1.0 // 爬虫
	// github.com/go-logr/stdr 1.2.2 // logr 使用stdr
	// github.com/go-logr/zerologr v1.2.3 // logr 使用zerologr
	// github.com/go-logr/glogr v1.2.2
	gorm.io/gorm v1.30.0
	gorm.io/plugin/opentelemetry v0.1.16
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/ClickHouse/ch-go v0.61.5 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.30.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.1 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20241020182519-7843d2ba8fdf // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/purego v0.8.1 // indirect
	github.com/elastic/gosigar v0.14.3 // indirect
	github.com/facebookgo/atomicfile v0.0.0-20151019160806-2de1f203e7d5 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/flynn/noise v1.1.0 // indirect
	github.com/francoispqt/gojay v1.2.13 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-resty/resty/v2 v2.13.1 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/pprof v0.0.0-20250208200701-d0013a598941 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/ipfs/boxo v0.27.2 // indirect
	github.com/ipfs/go-datastore v0.6.0 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/ipld/go-ipld-prime v0.21.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jbenet/go-temp-err-catcher v0.1.0 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/koron/go-ssdp v0.0.5 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-cidranger v1.1.0 // indirect
	github.com/libp2p/go-flow-metrics v0.2.0 // indirect
	github.com/libp2p/go-libp2p-asn-util v0.4.1 // indirect
	github.com/libp2p/go-libp2p-kbucket v0.6.4 // indirect
	github.com/libp2p/go-libp2p-record v0.3.1 // indirect
	github.com/libp2p/go-libp2p-routing-helpers v0.7.4 // indirect
	github.com/libp2p/go-msgio v0.3.0 // indirect
	github.com/libp2p/go-nat v0.2.0 // indirect
	github.com/libp2p/go-netroute v0.2.2 // indirect
	github.com/libp2p/go-reuseport v0.4.0 // indirect
	github.com/libp2p/go-yamux/v5 v5.0.0 // indirect
	github.com/libp2p/zeroconf/v2 v2.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/marten-seemann/tcp v0.0.0-20210406111302-dfbc87cc63fd // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/miekg/dns v1.1.63 // indirect
	github.com/mikioh/tcpinfo v0.0.0-20190314235526-30a79bb1804b // indirect
	github.com/mikioh/tcpopt v0.0.0-20190314235656-172688c1accc // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr-dns v0.4.1 // indirect
	github.com/multiformats/go-multiaddr-fmt v0.1.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.6.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/ginkgo/v2 v2.22.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pion/datachannel v1.5.10 // indirect
	github.com/pion/dtls/v2 v2.2.12 // indirect
	github.com/pion/dtls/v3 v3.0.6 // indirect
	github.com/pion/ice/v4 v4.0.10 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.15 // indirect
	github.com/pion/rtp v1.8.15 // indirect
	github.com/pion/sctp v1.8.39 // indirect
	github.com/pion/sdp/v3 v3.0.11 // indirect
	github.com/pion/srtp/v3 v3.0.4 // indirect
	github.com/pion/stun v0.6.1 // indirect
	github.com/pion/stun/v3 v3.0.0 // indirect
	github.com/pion/transport/v2 v2.2.10 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/turn/v4 v4.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/polydawn/refmt v0.89.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/quic-go v0.49.0 // indirect
	github.com/quic-go/webtransport-go v0.8.1-0.20241018022711-4ac2c9250e66 // indirect
	github.com/raulk/go-watchdog v1.3.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shirou/gopsutil/v4 v4.24.10 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/whyrusleeping/go-keyspace v0.0.0-20160322163242-5b898ac5add1 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.56.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/fx v1.23.0 // indirect
	go.uber.org/mock v0.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.14.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/exp v0.0.0-20250210185358-939b2ce775ac // indirect
	golang.org/x/image v0.23.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/clickhouse v0.7.0 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1 // indirect
	xorm.io/xorm v1.3.6 // indirect
)
