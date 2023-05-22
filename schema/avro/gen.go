package avro

//go:generate $GOPATH/bin/avrogo expense.avsc
//go:generate $GOPATH/bin/avrogo payment.avsc
//go:generate $GOPATH/bin/avrogo transaction.avsc
//go:generate $GOPATH/bin/avrogo order.avsc

// go install github.com/heetch/avro/cmd/avrogo...@latest
