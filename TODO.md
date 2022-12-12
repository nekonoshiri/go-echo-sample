# Body のログ

Request/response body もアクセスログと同じフォーマットで出力したい。
echo のログだけでは難しそうなので、`RequestLoggerWithConfig` と 3rd party の
ロギングライブラリを組み合わせる必要がありそう。

また、リクエスト ID（リクエスト毎に一意な ID）で追跡できるようにしたい。

ユースケースによっては出力したくないこともあるかもしれない（サイズが大きいなど）ので、
`BodyDump` 等ではなくユースケース側でログを吐く方がいいかもしれない？

3rd party のロギングライブラリを使用する場合、アプリケーションで使用しているログは
それに統一する。

# 統合テスト

TODO

# docker-compose.yaml を dev 用と prod 用にわける？

`docker-compose.dev.yaml`, `docker-compose.prod.yaml` みたいにする？

あと、MongoDB の username/password もベタ書きしなくていいようにしたい。

# Mongo URI の設定

Mongo の URI とかをたとえばローカルなら `localhost`,
Docker 使用時は `mongo` とかで切り替えられるようにする。

Viper を使うのがたぶん良い。
