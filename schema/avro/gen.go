package avro

//go:generate $GOPATH/bin/avrogen -pkg gen -o expense_gen.go -tags json:snake,avro:snake expense.avsc
//go:generate $GOPATH/bin/avrogen -pkg gen -o payment_gen.go -tags json:snake,avro:snake payment.avsc
//go:generate $GOPATH/bin/avrogen -pkg gen -o transaction_gen.go -tags json:snake,avro:snake transaction.avsc
//go:generate $GOPATH/bin/avrogen -pkg gen -o order_gen.go -tags json:snake,avro:snake order.avsc

// go install github.com/hamba/avro/v2/cmd/avrogen@latest
