package highheath

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

var expectedFileContent = `---
author: Alex Williams
date: 2021-01-01 00:00:00
title: "Comment from Alex Williams | 2021-01-01"
---
this is a great comment

`

func TestGetFileContent(t *testing.T) {
	g := NewWithT(t)
	comment := Comment{
		Contact: Contact{
			Name:    "Alex Williams",
			Message: "this is a great comment",
		},
		Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	actual := comment.GetFileContent()
	g.Expect(string(actual)).To(Equal(expectedFileContent))
}
