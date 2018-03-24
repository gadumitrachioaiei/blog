package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	pb "github.com/gadumitrachioaiei/blog"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

var lessons = []*pb.Lesson{
	{
		Domain:  "GO",
		Title:   "Lesson 0",
		Reads:   10,
		Created: ptypes.TimestampNow(),
		Content: "Lesson 0 is very important",
	},
	{
		Domain:  "GO",
		Title:   "Lesson 1",
		Reads:   11,
		Created: ptypes.TimestampNow(),
		Content: "Lesson 1 is very important",
	},
	{
		Domain:  "GO",
		Title:   "Lesson 2",
		Reads:   12,
		Created: ptypes.TimestampNow(),
		Content: "Lesson 2 is very important",
	},
}

func main() {
	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	server := grpc.NewServer()
	defer server.Stop()
	pb.RegisterBlogServer(server, &Blog{lessons: lessons})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Starting grpc server: %v", err)
	}
}

type Blog struct {
	mu      sync.Mutex
	lessons []*pb.Lesson
}

func (b *Blog) Search(ctx context.Context, query *pb.Query) (*pb.Content, error) {
	var result []*pb.Lesson
	term := query.GetTerm()
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, lesson := range b.lessons {
		if strings.Contains(lesson.Content, term) {
			result = append(result, lesson)
		}
	}
	return &pb.Content{Lessons: result}, nil
}

func (b *Blog) SearchAsStream(query *pb.Query, stream pb.Blog_SearchAsStreamServer) error {
	term := query.GetTerm()
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, lesson := range b.lessons {
		if strings.Contains(lesson.Content, term) {
			if err := stream.Send(lesson); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Blog) Write(stream pb.Blog_WriteServer) error {
	var content string
	for {
		p, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		content += p.Content
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	lesson := &pb.Lesson{
		Domain:  "GO",
		Title:   fmt.Sprintf("Lesson %d", len(b.lessons)),
		Content: content,
		Created: ptypes.TimestampNow(),
	}
	b.lessons = append(b.lessons, lesson)
	return stream.SendAndClose(lesson)
}

func (b *Blog) WriteRead(stream pb.Blog_WriteReadServer) error {
	var content string
	for {
		p, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		content += p.Content
		term := p.Content
		if len(p.Content) > 10 {
			term = p.Content[:10]
		}
		b.mu.Lock()
		for _, lesson := range b.lessons {
			if strings.Contains(lesson.Content, term) {
				if err := stream.Send(lesson); err != nil {
					return err
				}
			}
		}
		b.mu.Unlock()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lessons = append(b.lessons, &pb.Lesson{
		Domain:  "GO",
		Title:   fmt.Sprintf("Lesson %d", len(b.lessons)),
		Content: content,
		Created: ptypes.TimestampNow(),
	})
	return nil
}
