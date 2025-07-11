package metricsstorage

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

//--------------------producer--------------------

type Producer struct {
	ostream *os.File
	writer  *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	ostream, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		ostream: ostream,
		writer:  bufio.NewWriter(ostream),
	}, nil
}

func (p *Producer) Close() error {
	return p.ostream.Close()
}

func (p *Producer) WriteStorage() error {
	if _, err := p.ostream.Seek(0, 0); err != nil {
		return err
	}

	data, err := json.Marshal(&MS)
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}
	return p.writer.Flush()
}

func (p *Producer) RunSavingStorageRoutine(interval int) {
	go func() {
		for {
			if err := p.WriteStorage(); err != nil {
				log.Fatal().Err(err)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
}

//--------------------concumer--------------------

type Consumer struct {
	istream *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	istream, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		istream: istream,
		scanner: bufio.NewScanner(istream),
	}, nil
}

func (c *Consumer) ReadEvent() error {
	if !c.scanner.Scan() {
		return c.scanner.Err()
	}
	data := c.scanner.Bytes()

	if err := json.Unmarshal(data, &MS); err != nil {
		return err
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.istream.Close()
}
