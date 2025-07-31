# Logging helper

## Basic
There are four levels of logs:
* INFO:
  Normal processing log message. Sink: `stdout`
* ERROR:
  Recoverable error log message. Sink: `stderr`
* DEBUG:
  Misc log message. Only enable when debug mode is on. Sink: `stdout`
* FATAL:
  Unrecoverable error log message. It will panic the application. It should only be triggered for any error caused by **Non-user input**. Sink: `stderr`

## Methods
Multiple logging interfaces are provided in this `lg` package:
* `{{Level}}(v ...interface{})`: The simplest method to log one or more strings/variables. 
* `{{Level}}f(format string, v ...interface{})`: Format the log with c-style printf format.
* `{{Level}}c(ctx context.Context, format string, v ...interface{})`: This method accept a log context to inherit parameters from parent `With()` output.
* `With(ctx context.Context, format string, v ...interface{})`: This method writes no actual log. Instead, it returns a log context for ancestor application.

## Best practice
It's best for log analyser, like Kibana/ElasticSearch, to process structure data [Ref](https://stackify.com/what-is-structured-logging-and-why-developers-need-it/).

You are suggested to use `Xxxc()` series( `Infoc`, `Errorc` etc.) logging method to write log. The format string should compose multiple pairs of key/value paramters, like
`name=%s price=%.2f`, with `=` between key and value descriptor. 

Example:
```
func Somefunc(ctx context.Context, in *pb.Request) (*pb.Reply, error) {
    ctx = lg.With(ctx, "board_id=%s user_id=%s", in.GetBoardId(), in.GetUserId())

    // Do actual logic ...
    lg.Infoc(ctx, "Done")
    // Output:
    // [INFO]2019/07/02 07:25:51 Done board_id=12345 user_id=67890
}
```

## Timezone
*Beware!* The log record will be logged in **UTC** timezone, instead of China +0800. This is designed to fit our applications into datacenters in different timezones.

## Production deployment
(No mandatory right now) To deploy applications to production environment, it would be better to set JSON output encoding with `lg.SetEncoding(lg.TypeJSON)`, so that the filebeat can automatically collected the structured log, and help analysis.