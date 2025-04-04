package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"assignment2/shared"

	"github.com/streadway/amqp"
)

type PaperServer struct {
	papers map[string]shared.Paper
	nextID int
}

func NewPaperServer() *PaperServer {
	return &PaperServer{
		papers: make(map[string]shared.Paper),
		nextID: 1,
	}
}

func (ps *PaperServer) AddPaper(args shared.AddPaperArgs, reply *shared.AddPaperReply) error {
	paperID := fmt.Sprintf("P%d", ps.nextID)
	ps.nextID++

	ps.papers[paperID] = shared.Paper{
		Number:  paperID,
		Author:  args.Author,
		Title:   args.Title,
		Format:  args.Format,
		Content: args.Content,
	}

	reply.PaperNumber = paperID

	err := sendNotification(fmt.Sprintf("New paper added: %s by %s", args.Title, args.Author))
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
	}
	return nil
}

func (ps *PaperServer) ListPapers(args shared.ListPapersArgs, reply *shared.ListPapersReply) error {
	for _, paper := range ps.papers {
		reply.Papers = append(reply.Papers, struct {
			Number string
			Author string
			Title  string
		}{
			Number: paper.Number,
			Author: paper.Author,
			Title:  paper.Title,
		})
	}
	return nil
}

func (ps *PaperServer) GetPaperDetails(args shared.GetPaperArgs, reply *shared.GetPaperDetailsReply) error {
	paper, exists := ps.papers[args.PaperNumber]
	if !exists {
		return fmt.Errorf("paper not found")
	}

	reply.Author = paper.Author
	reply.Title = paper.Title
	return nil
}

func (ps *PaperServer) FetchPaperContent(args shared.FetchPaperArgs, reply *shared.FetchPaperReply) error {
	paper, exists := ps.papers[args.PaperNumber]
	if !exists {
		return fmt.Errorf("paper not found")
	}

	reply.Content = paper.Content
	reply.Format = paper.Format
	return nil
}

func sendNotification(message string) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"papers",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %v", err)
	}

	err = ch.Publish(
		"papers",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("Notification sent: %s", message)
	return nil
}

func main() {
	server := rpc.NewServer()
	ps := NewPaperServer()
	err := server.RegisterName("PaperServer", ps)
	if err != nil {
		log.Fatalf("Failed to register RPC server: %v", err)
	}

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Server is running on port 1234...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}
		go server.ServeConn(conn)
	}
}
