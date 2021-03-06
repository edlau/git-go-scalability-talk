package benchmark

import (
	"context"
	"fmt"
	"github.com/jadekler/git-go-scalability-talk/application/inputters"
	"github.com/jadekler/git-go-scalability-talk/application/model"
	"github.com/jadekler/git-go-scalability-talk/application/queues"
	"github.com/jadekler/git-go-scalability-talk/benchmark"
	"google.golang.org/grpc"
	"sync"
	"testing"
)

var ug unaryGrpcListenerBenchmark = unaryGrpcListenerBenchmark{}

type unaryGrpcListenerBenchmark struct {
	l  *listeners.UnaryGrpcListener
	wg *sync.WaitGroup
	q  queues.Queue
	p  int
	c  model.GrpcUnaryInputterServiceClient
}

func BenchmarkUnaryGrpcListener(b *testing.B) {
	if ug.l == nil {
		ug.p = benchmark.GetOpenTcpPort()

		ug.wg = &sync.WaitGroup{}
		ug.q = benchmark.NewWaitingQueue(ug.wg)

		ug.l = listeners.NewUnaryGrpcListener(ug.p)
		go ug.l.StartAccepting(ug.q)

		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", ug.p), grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		ug.c = model.NewGrpcUnaryInputterServiceClient(conn)
	}

	for i := 0; i < b.N; i++ {
		ug.wg.Add(1)

		_, err := ug.c.MakeRequest(context.Background(), &model.Request{Message: benchmark.VERY_LARGE_MESSAGE})
		if err != nil {
			panic(err)
		}
	}

	ug.wg.Wait()
}

func BenchmarkUnaryGrpcListenerParallel(b *testing.B) {
	if ug.l == nil {
		ug.p = benchmark.GetOpenTcpPort()

		ug.wg = &sync.WaitGroup{}
		ug.q = benchmark.NewWaitingQueue(ug.wg)

		ug.l = listeners.NewUnaryGrpcListener(ug.p)
		go ug.l.StartAccepting(ug.q)

		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", ug.p), grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		ug.c = model.NewGrpcUnaryInputterServiceClient(conn)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ug.wg.Add(1)

			_, err := ug.c.MakeRequest(context.Background(), &model.Request{Message: benchmark.SMALL_MESSAGE})
			if err != nil {
				panic(err)
			}
		}
	})

	ug.wg.Wait()
}
