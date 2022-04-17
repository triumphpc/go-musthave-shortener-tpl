// Test doing on start up server on 3200 port
package grpcshortener

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	proto "github.com/triumphpc/go-musthave-shortener-tpl/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	server := New(zap.NewNop(), rep, &sql.DB{}, &worker.Pool{})

	assert.IsType(t, &ShortenerServer{}, server)

}

func TestShortenerServer_AddLink(t *testing.T) {
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)

	rand.Seed(time.Now().UnixNano())

	links := []*proto.Link{
		{Link: `http://test1.ru` + strconv.Itoa(rand.Intn(99))},
		{Link: `http://test2.ru` + strconv.Itoa(rand.Intn(99))},
		{Link: `http://test3.ru` + strconv.Itoa(rand.Intn(99))},
	}

	for _, link := range links {
		// добавляем пользователей
		resp, err := c.AddLink(context.Background(), &proto.AddLinkRequest{
			Link: link,
		})
		if err != nil {
			log.Fatal(err)
		}

		assert.Equal(t, http.StatusCreated, int(resp.Code))

		// Пробуем повторно сохранить такую же
		resp, err = c.AddLink(context.Background(), &proto.AddLinkRequest{
			Link: link,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(resp)

		assert.Equal(t, http.StatusConflict, int(resp.Code))
	}
}

func TestShortenerServer_Ping(t *testing.T) {
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)

	resp, err := c.Ping(context.Background(), &proto.PingRequest{})
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, int(resp.Code))
}

func TestShortenerServer_AddBatch(t *testing.T) {
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)

	rand.Seed(time.Now().UnixNano())

	randId := strconv.Itoa(rand.Intn(99))
	randId2 := strconv.Itoa(rand.Intn(99))

	links := []*proto.JSONBatchLink{
		{Link: "http://test1.ru" + randId, Id: &proto.LinkID{Id: "id_" + randId}},
		{Link: "http://test1.ru" + randId2, Id: &proto.LinkID{Id: "id_" + randId2}},
	}

	resp, err := c.AddBatch(context.Background(), &proto.AddBatchRequest{Links: links})
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusCreated, int(resp.GetCode()))
	assert.Equal(t, len(links), len(resp.GetLinks()))

}

func TestShortenerServer_AddJSONLink(t *testing.T) {
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)

	rand.Seed(time.Now().UnixNano())

	link := proto.JSONLink{
		Link: "http://link.ru" + strconv.Itoa(rand.Intn(99)),
	}

	resp, err := c.AddJSONLink(context.Background(), &proto.AddJSONLinkRequest{Link: &link})
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusCreated, int(resp.GetCode()))
	assert.True(t, len(resp.GetLink().Link) > 0)

}

func TestShortenerServer_UserLinks(t *testing.T) {
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)

	resp, err := c.UserLinks(context.Background(), &proto.JSONUserLinksRequest{})
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, int(resp.GetCode()))
}

func TestShortenerServer_Stats(t *testing.T) {
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)

	resp, err := c.Stats(context.Background(), &proto.StatsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	assert.True(t, resp.Urls > 0)
	assert.True(t, resp.Users > 0)

}

func TestShortenerServer_Delete(t *testing.T) {
	// Registration
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)
	rand.Seed(time.Now().UnixNano())
	randId := strconv.Itoa(rand.Intn(99))

	links := []*proto.JSONBatchLink{
		{Link: "http://test1.ru" + randId, Id: &proto.LinkID{Id: "id_" + randId}},
	}

	resp, err := c.AddBatch(context.Background(), &proto.AddBatchRequest{Links: links})
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusCreated, int(resp.GetCode()))

	// And check delete handler
	deleteResp, err := c.Delete(context.Background(), &proto.DeleteRequest{
		Id: []*proto.LinkID{
			{Id: "id_" + randId},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusAccepted, int(deleteResp.GetCode()))
}

func TestShortenerServer_Origin(t *testing.T) {
	conn, err := grpc.Dial(`:3200`, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Interface type
	c := proto.NewShortenerClient(conn)

	rand.Seed(time.Now().UnixNano())

	link := proto.Link{Link: `http://test1.ru` + strconv.Itoa(rand.Intn(99))}

	// Add test link
	resp, err := c.AddLink(context.Background(), &proto.AddLinkRequest{
		Link: &link,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Get short link
	shortLink := resp.GetLink().Link

	p := strings.Split(shortLink, "/")
	shortLink = p[len(p)-1]

	// Get origin from short
	respOrigin, err := c.Origin(context.Background(), &proto.OriginRequest{
		Link: &proto.ShortLink{
			Link: shortLink,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, respOrigin.Link.Link, link.Link)

}
