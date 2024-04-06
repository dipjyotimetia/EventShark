package main

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	topic := "expense-topic"
	ctx := context.Background()

	seeds := []string{"localhost:9092"}
	opts := []kgo.Opt{}
	opts = append(opts,
		kgo.SeedBrokers(seeds...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	for {
		fetches := client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			panic(fmt.Sprint(errs))
		}
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			topicInfo := fmt.Sprintf("topic: %s (%d|%d)",
				record.Topic, record.Partition, record.Offset)
			messageInfo := fmt.Sprintf("key: %s, Value: %s",
				record.Key, record.Value)
			fmt.Printf("Message consumed: %s, %s \n", topicInfo, messageInfo)
		}
	}
}
