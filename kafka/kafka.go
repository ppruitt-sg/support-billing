/**********************************/
/*    Only works on localhost    */
/**********************************/

package kafka

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Shopify/sarama"
)

var topic = "sb-alerts"

func WriteError(s string, e error) {
	// Create a producer
	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	defer producer.Close()

	// Create JSON
	value, err := createJSON(s, e)
	if err != nil {
		panic(err)
	}

	//Create Message
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	// Input message"
	producer.Input() <- message
}

func createJSON(s string, err error) ([]byte, error) {
	j := struct {
		Timestamp int64
		Action    string
		Error     string
	}{time.Now().Unix(), s, err.Error()}

	return json.Marshal(j)
}
