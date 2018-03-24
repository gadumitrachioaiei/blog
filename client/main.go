package main

import (
	"fmt"
	"io"
	"log"

	"golang.org/x/net/context"

	"github.com/gadumitrachioaiei/blog"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":5050", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}
	defer conn.Close()
	client := blog.NewBlogClient(conn)
	// search
	Search(client)
	// search as a stream
	SearchStream(client)
	// write a new lesson
	Write(client)
	// write a new lesson and receive similar lessons
	WriteRead(client)
}

func Search(client blog.BlogClient) {
	lessons, err := client.Search(context.Background(), &blog.Query{Term: "Lesson"})
	if err != nil {
		log.Fatalf("Cannot query lessons: %v", err)
	}
	fmt.Println("lessons are", lessons)
}

func SearchStream(client blog.BlogClient) {
	searchStream, err := client.SearchAsStream(context.Background(), &blog.Query{Term: "Lesson"})
	if err != nil {
		log.Fatalf("search as stream: %v", err)
	}
	for {
		lesson, err := searchStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("search as stream receive: %v", err)
		}
		fmt.Println("received lesson", lesson)
	}
}

func Write(client blog.BlogClient) {
	writeStream, err := client.Write(context.Background())
	if err != nil {
		log.Fatalf("write stream: %v", err)
	}
	for i := 0; i < 10; i++ {
		if err := writeStream.Send(&blog.Paragraph{Content: "Lesson "}); err != nil {
			log.Fatalf("write stream send: %v", err)
		}
	}
	if lesson, err := writeStream.CloseAndRecv(); err != nil {
		log.Fatalf("write stream receive bla: %v", err)
	} else {
		fmt.Println("written lesson", lesson)
	}
}

func WriteRead(client blog.BlogClient) {
	writeReadStream, err := client.WriteRead(context.Background())
	if err != nil {
		log.Fatalf("writeReadStream: %v", err)
	}
	readChan := make(chan bool)
	go func() {
		for {
			lesson, err := writeReadStream.Recv()
			if err == io.EOF {
				close(readChan)
				break
			}
			if err != nil {
				log.Fatalf("writeReadStream receive: %v", err)
			}
			fmt.Println("reading similar lesson", lesson)
		}
	}()
	for i := 0; i < 10; i++ {
		if err := writeReadStream.Send(&blog.Paragraph{Content: "Lesson write read "}); err != nil {
			log.Fatalf("writeReadStream send: %v", err)
		}
	}
	writeReadStream.CloseSend()
	<-readChan
}
