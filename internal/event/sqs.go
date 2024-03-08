package event

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	log "github.com/sirupsen/logrus"
)

type Sqs struct {
	QueueUrl string `json:"queueUrl"`
	Uri      string

	client *sqs.SQS `json:"-"`
}

type SQSMessage struct {
	Records []struct {
		EventVersion string    `json:"eventVersion"`
		EventSource  string    `json:"eventSource"`
		AwsRegion    string    `json:"awsRegion"`
		EventTime    time.Time `json:"eventTime"`
		EventName    string    `json:"eventName"`
		UserIdentity struct {
			PrincipalID string `json:"principalId"`
		} `json:"userIdentity"`
		RequestParameters struct {
			SourceIPAddress string `json:"sourceIPAddress"`
		} `json:"requestParameters"`
		ResponseElements struct {
			XAmzRequestID string `json:"x-amz-request-id"`
			XAmzID2       string `json:"x-amz-id-2"`
		} `json:"responseElements"`
		S3 struct {
			S3SchemaVersion string `json:"s3SchemaVersion"`
			ConfigurationID string `json:"configurationId"`
			Bucket          struct {
				Name          string `json:"name"`
				OwnerIdentity struct {
					PrincipalID string `json:"principalId"`
				} `json:"ownerIdentity"`
				Arn string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key       string `json:"key"`
				Size      int    `json:"size"`
				ETag      string `json:"eTag"`
				Sequencer string `json:"sequencer"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func (s *Sqs) Init(cfg string) error {
	s.QueueUrl = cfg
	sess := session.Must(session.NewSession())
	if AssumeRoleArn != "" {
		// Assume the specified IAM role
		creds := stscreds.NewCredentials(sess, AssumeRoleArn)
		// Create a new S3 service client with the assumed role credentials
		s.client = sqs.New(sess, &aws.Config{Credentials: creds})
	} else {
		// Create a new S3 service client
		s.client = sqs.New(sess)
	}
	return nil
}

func (s *Sqs) GetUri() (string, error) {
	l := log.WithFields(log.Fields{
		"fn": "Sqs.GetUri",
	})
	l.Debug("getting URI from SQS")
	var an []*string
	// assume some filtering would be done
	an = append(an, aws.String("All"))
	var man []*string
	man = append(man, aws.String("All"))
	rmi := &sqs.ReceiveMessageInput{
		// set queue URL
		QueueUrl:       aws.String(s.QueueUrl),
		AttributeNames: an,
		// retrieve all
		MessageAttributeNames: man,
		// retrieve one message at a time
		MaxNumberOfMessages: aws.Int64(1),
		// do not timeout visibility - for testing
		//VisibilityTimeout: aws.Int64(0),
	}
	m, err := s.client.ReceiveMessage(rmi)
	if err != nil {
		l.WithError(err).Error("error receiving message")
		return "", err
	}
	if len(m.Messages) == 0 {
		l.Debug("no messages")
		return "", nil
	}
	sm := &SQSMessage{}
	if err := json.Unmarshal([]byte(*m.Messages[0].Body), sm); err != nil {
		l.WithError(err).Error("error unmarshalling message")
		return "", err
	}
	l.Debugf("message: %+v", *m.Messages[0].Body)
	// ensure eventName is ObjectCreated:*
	if !strings.HasPrefix(sm.Records[0].EventName, "ObjectCreated:") {
		l.Debug("not an ObjectCreated event")
		return "", nil
	}
	// set the URI to the object
	s.Uri = "s3://" + sm.Records[0].S3.Bucket.Name + "/" + sm.Records[0].S3.Object.Key
	// delete the message
	dmi := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.QueueUrl),
		ReceiptHandle: m.Messages[0].ReceiptHandle,
	}
	if _, err := s.client.DeleteMessage(dmi); err != nil {
		l.WithError(err).Error("error deleting message")
		return "", err
	}
	return s.Uri, nil
}
