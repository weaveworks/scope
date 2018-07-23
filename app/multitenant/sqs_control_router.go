package multitenant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/weaveworks/common/instrument"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/xfer"
)

var (
	longPollTime       = aws.Int64(10)
	sqsRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "sqs_request_duration_seconds",
		Help:      "Time in seconds spent doing SQS requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status_code"})
)

func init() {
	prometheus.MustRegister(sqsRequestDuration)
}

// sqsControlRouter:
// Creates a queue for every probe that connects to it, and a queue for
// responses back to it.  When it receives a request, posts it to the
// probe queue.  When probe receives a request, handles it and posts the
// response back to the response queue.
type sqsControlRouter struct {
	service          *sqs.SQS
	responseQueueURL *string
	userIDer         UserIDer
	prefix           string
	rpcTimeout       time.Duration

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
func NewSQSControlRouter(config *aws.Config, userIDer UserIDer, prefix string, rpcTimeout time.Duration) app.ControlRouter {
	result := &sqsControlRouter{
		service:          sqs.New(session.New(config)),
		responseQueueURL: nil,
		userIDer:         userIDer,
		prefix:           prefix,
		rpcTimeout:       rpcTimeout,
		responses:        map[string]chan xfer.Response{},
		probeWorkers:     map[int64]*probeWorker{},
	}
	go result.loop()
	return result
}

func (cr *sqsControlRouter) Stop() error {
	return nil
}

func (cr *sqsControlRouter) setResponseQueueURL(url *string) {
	cr.mtx.Lock()
	defer cr.mtx.Unlock()
	cr.responseQueueURL = url
}

func (cr *sqsControlRouter) getResponseQueueURL() *string {
	cr.mtx.Lock()
	defer cr.mtx.Unlock()
	return cr.responseQueueURL
}

func (cr *sqsControlRouter) getOrCreateQueue(ctx context.Context, name string) (*string, error) {
	// CreateQueue creates a queue or if it already exists, returns url of said queue
	var createQueueRes *sqs.CreateQueueOutput
	var err error
	err = instrument.TimeRequestHistogram(ctx, "SQS.CreateQueue", sqsRequestDuration, func(_ context.Context) error {
		createQueueRes, err = cr.service.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(name),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return createQueueRes.QueueUrl, nil
}

func (cr *sqsControlRouter) loop() {
	var (
		responseQueueURL *string
		err              error
		ctx              = context.Background()
	)
	for {
		// This app has a random id and uses this as a return path for all responses from probes.
		name := fmt.Sprintf("%scontrol-app-%d", cr.prefix, rand.Int63())
		responseQueueURL, err = cr.getOrCreateQueue(ctx, name)
		if err != nil {
			log.Errorf("Failed to create queue: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		cr.setResponseQueueURL(responseQueueURL)
		break
	}

	for {
		var res *sqs.ReceiveMessageOutput
		var err error
		err = instrument.TimeRequestHistogram(ctx, "SQS.ReceiveMessage", sqsRequestDuration, func(_ context.Context) error {
			res, err = cr.service.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:        responseQueueURL,
				WaitTimeSeconds: longPollTime,
			})
			return err
		})
		if err != nil {
			log.Errorf("Error receiving message from %s: %v", *responseQueueURL, err)
			continue
		}

		if len(res.Messages) == 0 {
			continue
		}
		if err := cr.deleteMessages(ctx, responseQueueURL, res.Messages); err != nil {
			log.Errorf("Error deleting message from %s: %v", *responseQueueURL, err)
		}
		cr.handleResponses(res)
	}
}

func (cr *sqsControlRouter) deleteMessages(ctx context.Context, queueURL *string, messages []*sqs.Message) error {
	entries := []*sqs.DeleteMessageBatchRequestEntry{}
	for _, message := range messages {
		entries = append(entries, &sqs.DeleteMessageBatchRequestEntry{
			ReceiptHandle: message.ReceiptHandle,
			Id:            message.MessageId,
		})
	}
	return instrument.TimeRequestHistogram(ctx, "SQS.DeleteMessageBatch", sqsRequestDuration, func(_ context.Context) error {
		_, err := cr.service.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
			QueueUrl: queueURL,
			Entries:  entries,
		})
		return err
	})
}

func (cr *sqsControlRouter) handleResponses(res *sqs.ReceiveMessageOutput) {
	cr.mtx.Lock()
	defer cr.mtx.Unlock()

	for _, message := range res.Messages {
		var sqsResponse sqsResponseMessage
		if err := json.NewDecoder(bytes.NewBufferString(*message.Body)).Decode(&sqsResponse); err != nil {
			log.Errorf("Error decoding message: %v", err)
			continue
		}

		waiter, ok := cr.responses[sqsResponse.ID]
		if !ok {
			log.Errorf("Dropping response %s - no one waiting for it!", sqsResponse.ID)
			continue
		}
		waiter <- sqsResponse.Response
	}
}

func (cr *sqsControlRouter) sendMessage(ctx context.Context, queueURL *string, message interface{}) error {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(message); err != nil {
		return err
	}
	log.Debugf("sendMessage to %s: %s", *queueURL, buf.String())

	return instrument.TimeRequestHistogram(ctx, "SQS.SendMessage", sqsRequestDuration, func(_ context.Context) error {
		_, err := cr.service.SendMessage(&sqs.SendMessageInput{
			QueueUrl:    queueURL,
			MessageBody: aws.String(buf.String()),
		})
		return err
	})
}

func (cr *sqsControlRouter) Handle(ctx context.Context, probeID string, req xfer.Request) (xfer.Response, error) {
	// Make sure we know the users
	userID, err := cr.userIDer(ctx)
	if err != nil {
		return xfer.Response{}, err
	}

	// Get the queue url for the local (control app) queue, and for the probe.
	responseQueueURL := cr.getResponseQueueURL()
	if responseQueueURL == nil {
		return xfer.Response{}, fmt.Errorf("no SQS queue yet")
	}

	var probeQueueURL *sqs.GetQueueUrlOutput
	err = instrument.TimeRequestHistogram(ctx, "SQS.GetQueueUrl", sqsRequestDuration, func(_ context.Context) error {
		probeQueueName := fmt.Sprintf("%sprobe-%s-%s", cr.prefix, userID, probeID)
		probeQueueURL, err = cr.service.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: aws.String(probeQueueName),
		})
		return err
	})
	if err != nil {
		return xfer.Response{}, err
	}

	// Add a response channel before we send the request, to prevent races
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
	if err := instrument.TimeRequestHistogram(ctx, "SQS.SendMessage", sqsRequestDuration, func(ctx context.Context) error {
		return cr.sendMessage(ctx, probeQueueURL.QueueUrl, sqsRequestMessage{
			ID:               id,
			Request:          req,
			ResponseQueueURL: *responseQueueURL,
		})
	}); err != nil {
		return xfer.Response{}, err
	}

	// Finally, wait for a response on our queue
	select {
	case response := <-waiter:
		return response, nil
	case <-time.After(cr.rpcTimeout):
		return xfer.Response{}, fmt.Errorf("request timed out")
	}
}

func (cr *sqsControlRouter) Register(ctx context.Context, probeID string, handler xfer.ControlHandlerFunc) (int64, error) {
	userID, err := cr.userIDer(ctx)
	if err != nil {
		return 0, err
	}

	name := fmt.Sprintf("%sprobe-%s-%s", cr.prefix, userID, probeID)
	queueURL, err := cr.getOrCreateQueue(ctx, name)
	if err != nil {
		return 0, err
	}

	pwID := rand.Int63()
	pw := &probeWorker{
		ctx:             ctx,
		router:          cr,
		requestQueueURL: queueURL,
		handler:         handler,
		quit:            make(chan struct{}),
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

// a probeWorker encapsulates a goroutine serving a probe's websocket connection.
type probeWorker struct {
	ctx             context.Context
	router          *sqsControlRouter
	requestQueueURL *string
	handler         xfer.ControlHandlerFunc
	quit            chan struct{}
	done            sync.WaitGroup
}

func (pw *probeWorker) stop() {
	close(pw.quit)
	pw.done.Wait()
}

func (pw *probeWorker) loop() {
	defer pw.done.Done()
	for {
		// have we been stopped?
		select {
		case <-pw.quit:
			return
		default:
		}

		var res *sqs.ReceiveMessageOutput
		var err error
		err = instrument.TimeRequestHistogram(pw.ctx, "SQS.ReceiveMessage", sqsRequestDuration, func(_ context.Context) error {
			res, err = pw.router.service.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:        pw.requestQueueURL,
				WaitTimeSeconds: longPollTime,
			})
			return err
		})
		if err != nil {
			log.Errorf("Error receiving message: %v", err)
			continue
		}

		if len(res.Messages) == 0 {
			continue
		}
		if err := pw.router.deleteMessages(pw.ctx, pw.requestQueueURL, res.Messages); err != nil {
			log.Errorf("Error deleting message from %s: %v", *pw.requestQueueURL, err)
		}

		for _, message := range res.Messages {
			var sqsRequest sqsRequestMessage
			if err := json.NewDecoder(bytes.NewBufferString(*message.Body)).Decode(&sqsRequest); err != nil {
				log.Errorf("Error decoding message from: %v", err)
				continue
			}

			response := pw.handler(sqsRequest.Request)

			if err := pw.router.sendMessage(pw.ctx, &sqsRequest.ResponseQueueURL, sqsResponseMessage{
				ID:       sqsRequest.ID,
				Response: response,
			}); err != nil {
				log.Errorf("Error sending response: %v", err)
			}
		}
	}
}
