package kafka

import "github.com/confluentinc/confluent-kafka-go/kafka"

// kafkaMessageCarrier injects and extracts tracing headers from a kafka.Message.
type kafkaMessageCarrier struct {
	msg *kafka.Message
}

func (c kafkaMessageCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c kafkaMessageCarrier) Set(key, val string) {
	c.msg.Headers = append(c.msg.Headers, kafka.Header{Key: key, Value: []byte(val)})
}

func (c kafkaMessageCarrier) Keys() []string {
	keys := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		keys[i] = h.Key
	}
	return keys
}
