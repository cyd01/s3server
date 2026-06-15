package api

import (
	"encoding/xml"
	"time"
)

type ListBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

type Bucket struct {
	Name         string    `xml:"Name"`
	CreationDate time.Time `xml:"CreationDate"`
}

type ListBucketResult struct {
	XMLName  xml.Name `xml:"ListBucketResult"`
	Name     string   `xml:"Name"`
	KeyCount int      `xml:"KeyCount"`
	Contents []Object `xml:"Contents"`
}

type Object struct {
	Key          string    `xml:"Key"`
	Size         int64     `xml:"Size"`
	LastModified time.Time `xml:"LastModified"`
}
