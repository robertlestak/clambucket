package clambucket

import (
	"fmt"
	"os"
	"path"

	"github.com/robertlestak/clambucket/internal/clamav"
	"github.com/robertlestak/clambucket/internal/object"
	log "github.com/sirupsen/logrus"
)

var (
	Downstreams   *Downstream
	AssumeRoleArn string
)

type Downstream struct {
	Quarantine       string
	QuarantinePrefix string
	Clean            string
	CleanPrefix      string
}

type Object struct {
	Uri string
}

func (o *Object) moveToBucket(bucket string) error {
	l := log.WithFields(log.Fields{
		"fn":     "Object.moveToBucket",
		"object": o.Uri,
	})
	l.Debug("moving object to bucket")
	_, srckey, err := object.ParseS3Uri(o.Uri)
	if err != nil {
		return err
	}
	dstUri := fmt.Sprintf("s3://%s/%s", bucket, srckey)
	if err := object.S3Copy(o.Uri, dstUri, AssumeRoleArn); err != nil {
		return err
	}
	if err := object.S3Delete(o.Uri, AssumeRoleArn); err != nil {
		return err
	}
	return nil
}

func (o *Object) moveToQuarantine() error {
	l := log.WithFields(log.Fields{
		"fn":     "Object.moveToQuarantine",
		"object": o.Uri,
	})
	l.Debug("moving object to quarantine")
	return o.moveToBucket(path.Join(Downstreams.Quarantine, Downstreams.QuarantinePrefix))
}

func (o *Object) moveToClean() error {
	l := log.WithFields(log.Fields{
		"fn":     "Object.moveToClean",
		"object": o.Uri,
	})
	l.Debug("moving object to clean")
	return o.moveToBucket(path.Join(Downstreams.Clean, Downstreams.CleanPrefix))
}

func (o *Object) Scan() error {
	l := log.WithFields(log.Fields{
		"fn":     "ObjectScan.Scan",
		"object": o.Uri,
	})
	l.Debug("scanning object")
	if o.Uri == "" {
		return fmt.Errorf("URI is empty")
	}
	t, err := os.MkdirTemp("", "clambucket")
	if err != nil {
		return err
	}
	defer os.RemoveAll(t)
	if err := object.S3Get(o.Uri, t, AssumeRoleArn); err != nil {
		return err
	}
	summary, err := clamav.Scan(t)
	if err != nil {
		return err
	}
	if Downstreams == nil {
		l.Warn("no downstream buckets configured")
		l.Infof("scan summary: %d infected, %d scanned", summary.Infected, summary.Scanned)
		return nil
	}
	if summary.Infected > 0 {
		l.Warn("object is infected")
		if err := o.moveToQuarantine(); err != nil {
			return err
		}
	} else {
		l.Info("object is clean")
		if err := o.moveToClean(); err != nil {
			return err
		}
	}
	return nil
}

func Scan(uri string) error {
	o := &Object{
		Uri: uri,
	}
	return o.Scan()
}
