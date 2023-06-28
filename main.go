package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
	"strings"
	"sync"
)

// read a range, then remove read content
func rangeAndTrim(ctx context.Context, client *redis.Client, key string, end int64) (read []string, err error) {
	pipeline := client.Pipeline()
	cmd := pipeline.LRange(ctx, key, 0, end-1)
	pipeline.LTrim(ctx, key, end, -1)
	if _, err = pipeline.Exec(ctx); err != nil {
		return
	}
	read, err = cmd.Result()
	return
}

func produceNumberContent(ctx context.Context, client *redis.Client, key string, max, batch int) (err error) {
	values := make([]string, 0, batch)
	for i := 0; i < max; i++ {
		values = append(values, strconv.Itoa(i))
		if (i+1)%batch == 0 {
			if cmd := client.RPush(ctx, key, values); cmd.Err() != nil {
				err = cmd.Err()
				return
			}
			values = make([]string, 0, batch)
		}
	}
	if len(values) > 0 {
		cmd := client.RPush(ctx, key, values)
		err = cmd.Err()
	}
	return
}

func lineBreak(values []string, colum int) [][]string {
	lv := len(values)
	row := lv / colum
	if lv%colum != 0 {
		row++
	}
	result := make([][]string, 0, row)
	i := 0
	for ; i < lv; i += colum {
		if i+colum >= lv {
			result = append(result, values[i:])
			break
		}
		result = append(result, values[i:i+colum])
	}
	return result
}

const (
	keyName = "test"
)

func main() {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	ctx := context.Background()
	defer func() {
		left := client.LRange(ctx, keyName, 0, -1).Val()
		log.Printf("left: \n%#v", left)
		client.Del(ctx, keyName)
	}()
	max := 10000
	if err := produceNumberContent(ctx, client, keyName, max, 1000); err != nil {
		log.Printf("produce failed, %#v", err)
	}
	join := new(sync.WaitGroup)
	step := 100
	for i := 0; i < max; i += step {
		join.Add(1)
		go func() {
			defer join.Done()
			if read, err := rangeAndTrim(ctx, client, keyName, int64(step)); err != nil {
				log.Printf("range and trim failed, %#v", err)
			} else {
				result := lineBreak(read, 10)
				buffer := new(strings.Builder)
				for _, row := range result {
					buffer.WriteString(strings.Join(row, ", "))
					buffer.WriteString("\n")
				}
				log.Printf("read: \n%s", buffer)
			}
		}()
	}
	join.Wait()
}
