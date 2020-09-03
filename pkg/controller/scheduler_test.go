package controller_test

import (
	"github.com/k-kinzal/progressived/pkg/controller"
	"testing"
	"time"
)

func TestScheduler_Add(t *testing.T) {
	scheduler := controller.NewScheduler()
	t1, _ := time.Parse("2006-01-02", "2020-01-01")
	scheduler.Add(t1, func() {})
	t2, _ := time.Parse("2006-01-02", "2020-01-03")
	scheduler.Add(t2, func() {})
	t3, _ := time.Parse("2006-01-02", "2020-01-02")
	scheduler.Add(t3, func() {})
	t4, _ := time.Parse("2006-01-02", "2019-12-31")
	scheduler.Add(t4, func() {})
}
