package multitenant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/xfer"
)

var (
	longPollTime = aws.Int64(10)
	rpcTimeout   = time.Minute
)

// sqsControlRouter:
// Creates a queue for every probe that connects to it, and a queue for
// responses back to it.  When it recieves a request, posts it to the
// probe queue.  When probe recieves a request, handles it and posts the
// response back to the response queue.
type sqsControlRouter struct {
	service  *sqs.SQS
	queueURL *string
	userIDer UserIDer

	mtx          sync.Mutex
	responses    map[string]chan xfer.Response
	probeWorkers map[int64]*probeWorker
}

type sqsRequestMessage struct {
	ID               string
	Request          xfer.Request
	ResponseQueueURL string
}

type sqsResponseMessage struct {
	ID       string
	Response xfer.Response
}

// NewSQSControlRouter the harbinger of death
func NewSQSControlRouter(url, region string, creds *credentials.Credentials, userIDer UserIDer) app.ControlRouter {
	result := &sqsControlRouter{
		service: sqs.New(session.New(aws.NewConfig().
			WithEndpoint(url).
			WithRegion(region).
			WithCredentials(creds))),
		queueURL:     nil,
		userIDer:     userIDer,
		responses:    map[string]chan xfer.Response{},
		probeWorkers: map[int64]*probeWorker{},
	}
	go result.loop()
	return result
}

func (cr *sqsControlRouter) Stop() error {
	return nil
}

func (cr *sqsControlRouter) setQueueURL(url *string) {
	cr.mtx.Lock()
	defer cr.mtx.Unlock()
	cr.queueURL = url
}

func (cr *sqsControlRouter) getQueueURL() *string {
	cr.mtx.Lock()
	defer cr.mtx.Unlock()
	return cr.queueURL
}

func (cr *sqsControlRouter) getOrCreateQueue(name string) (*string, error) {
	getQueueURLRes, err := cr.service.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err == nil {
		return getQueueURLRes.QueueUrl, nil
	}

	createQueueRes, err := cr.service.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}
	return createQueueRes.QueueUrl, nil
}

func (cr *sqsControlRouter) loop() {
	for {
		// This app has a random id and uses this as a return path for all responses from probes.
		name := fmt.Sprintf("control-app-%d", rand.Int63())
		queueURL, err := cr.getOrCreateQueue(name)
		if err != nil {
			log.Errorf("Failed to create queue: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		cr.setQueueURL(queueURL)
		break
	}

	for {
		res, err := cr.service.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:        cr.queueURL,
			WaitTimeSeconds: longPollTime,
		})
		if err != nil {
			log.Errorf("Error recieving message from %s: %v", *cr.queueURL, err)
			continue
		}
		if len(res.Messages) == 0 {
			continue
		}
		cr.handleResponses(res)
		if err := cr.deleteMessages(cr.queueURL, res.Messages); err != nil {
			log.Errorf("Error deleting message from %s: %v", *cr.queueURL, err)
		}
	}
}

func (cr *sqsControlRouter) deleteMessages(queueURL *string, messages []*sqs.Message) error {
	entries := []*sqs.DeleteMessageBatchRequestEntry{}
	for _, message := range messages {
		entries = append(entries, &sqs.DeleteMessageBatchRequestEntry{
			ReceiptHandle: message.ReceiptHandle,
			Id:            message.MessageId,
		})
	}
	_, err := cr.service.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
		QueueUrl: queueURL,
		Entries:  entries,
	})
	return err
}

func (cr *sqsControlRouter) handleResponses(res *sqs.ReceiveMessageOutput) {
	sqsResponses := []sqsResponseMessage{}
	for _, message := range res.Messages {
		var sqsResponse sqsResponseMessage
		if err := json.NewDecoder(bytes.NewBufferString(*message.Body)).Decode(&sqsResponse); err != nil {
			log.Errorf("Error decoding message: %v", err)
			continue
		}
		sqsResponses = append(sqsResponses, sqsResponse)
	}

	for _, sqsResponse := range sqsResponses {
		cr.mtx.Lock()
		waiter, ok := cr.responses[sqsResponse.ID]
		cr.mtx.Unlock()

		if !ok {
			log.Errorf("Dropping response %s - no one waiting for it!", sqsResponse.ID)
			continue
		}
		waiter <- sqsResponse.Response
	}
}

func (cr *sqsControlRouter) sendMessage(queueURL *string, message interface{}) error {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(message); err != nil {
		return err
	}
	log.Infof("sendMessage to %s: %s", *queueURL, buf.String())
	_, err := cr.service.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    queueURL,
		MessageBody: aws.String(buf.String()),
	})
	return err
}

func (cr *sqsControlRouter) Handle(ctx context.Context, probeID string, req xfer.Request) (xfer.Response, error) {
	// Make sure we know the users
	userID, err := cr.userIDer(ctx)
	if err != nil {
		return xfer.Response{}, err
	}

	// Get the queue url for the local (control app) queue, and for the probe.
	queueURL := cr.getQueueURL()
	if queueURL == nil {
		return xfer.Response{}, fmt.Errorf("No SQS queue yet!")
	}
	probeQueueName := fmt.Sprintf("probe-%s-%s", userID, probeID)
	probeQueueURL, err := cr.service.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(probeQueueName),
	})
	if err != nil {
		return xfer.Response{}, err
	}

	// Wait for a response befor we send the request, to prevent races
	id := fmt.Sprintf("request-%s-%d", userID, rand.Int63())
	waiter := make(chan xfer.Response, 1)
	cr.mtx.Lock()
	cr.responses[id] = waiter
	cr.mtx.Unlock()
	defer func() {
		cr.mtx.Lock()
		delete(cr.responses, id)
		cr.mtx.Unlock()
	}()

	// Next, send the request to that queue
	if err := cr.sendMessage(probeQueueURL.QueueUrl, sqsRequestMessage{
		ID:               id,
		Request:          req,
		ResponseQueueURL: *queueURL,
	}); err != nil {
		return xfer.Response{}, err
	}

	// Finally, wait for a response on our queue
	select {
	case response := <-waiter:
		return response, nil
	case <-time.After(rpcTimeout):
		return xfer.Response{}, fmt.Errorf("Request timedout.")
	}
}

func (cr *sqsControlRouter) Register(ctx context.Context, probeID string, handler xfer.ControlHandlerFunc) (int64, error) {
	userID, err := cr.userIDer(ctx)
	if err != nil {
		return 0, err
	}

	name := fmt.Sprintf("probe-%s-%s", userID, probeID)
	queueURL, err := cr.getOrCreateQueue(name)
	if err != nil {
		return 0, err
	}

	pwID := rand.Int63()
	pw := &probeWorker{
		router:   cr,
		queueURL: queueURL,
		handler:  handler,
		quit:     make(chan struct{}),
	}
	pw.done.Add(1)
	go pw.loop()

	cr.mtx.Lock()
	defer cr.mtx.Unlock()
	cr.probeWorkers[pwID] = pw
	return pwID, nil
}

func (cr *sqsControlRouter) Deregister(_ context.Context, probeID string, id int64) error {
	cr.mtx.Lock()
	pw, ok := cr.probeWorkers[id]
	delete(cr.probeWorkers, id)
	cr.mtx.Unlock()
	if ok {
		pw.stop()
	}
	return nil
}

type probeWorker struct {
	router   *sqsControlRouter
	queueURL *string
	handler  xfer.ControlHandlerFunc
	quit     chan struct{}
	done     sync.WaitGroup
}

func (pw *probeWorker) stop() {
	close(pw.quit)
	pw.done.Wait()
}

func (pw *probeWorker) loop() {
	for {
		// have we been stopped?
		select {
		case <-pw.quit:
			return
		default:
		}

		res, err := pw.router.service.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:        pw.queueURL,
			WaitTimeSeconds: longPollTime,
		})
		if err != nil {
			log.Errorf("Error recieving message: %v", err)
			continue
		}
		if len(res.Messages) == 0 {
			continue
		}

		// TODO do we need to parallelise the handling of requests?
		for _, message := range res.Messages {
			var sqsRequest sqsRequestMessage
			if err := json.NewDecoder(bytes.NewBufferString(*message.Body)).Decode(&sqsRequest); err != nil {
				log.Errorf("Error decoding message from: %v", err)
				continue
			}

			if err := pw.router.sendMessage(&sqsRequest.ResponseQueueURL, sqsResponseMessage{
				ID:       sqsRequest.ID,
				Response: pw.handler(sqsRequest.Request),
			}); err != nil {
				log.Errorf("Error sending response: %v", err)
			}
		}

		if err := pw.router.deleteMessages(pw.queueURL, res.Messages); err != nil {
			log.Errorf("Error deleting message from %s: %v", *pw.queueURL, err)
		}
	}
}
