package queue

type Queue interface {
	Enqueue(requestID string, requestData interface{}) error
	Dequeue() (string, interface{}, error)
}
