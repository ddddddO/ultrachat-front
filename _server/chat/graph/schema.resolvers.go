package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.26

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/ddddddO/ultrachat-front/_server/chat/graph/model"
)

// SendMessage is the resolver for the sendMessage field.
func (r *mutationResolver) SendMessage(ctx context.Context, message string) (*model.ChatMessage, error) {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)

	chatMessage := &model.ChatMessage{
		ID: ulid.MustNew(ulid.Timestamp(t), entropy).String(), // FIXME: return err
		// User:      user,
		Message:   message,
		CreatedAt: time.Now().UTC().String(),
	}

	// 投稿されたメッセージを保存し、subscribeしている全てのコネクションにブロードキャスト
	r.mutex.Lock()
	r.messages = append(r.messages, chatMessage)
	for _, ch := range r.subscribers {
		ch <- chatMessage
	}
	r.mutex.Unlock()

	return chatMessage, nil
}

// GetChatMessages is the resolver for the getChatMessages field.
func (r *queryResolver) GetChatMessages(ctx context.Context) ([]*model.ChatMessage, error) {
	return r.messages, nil
}

// MessageSent is the resolver for the messageSent field.
func (r *subscriptionResolver) MessageSent(ctx context.Context) (<-chan *model.ChatMessage, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	user := fmt.Sprintf("匿名希望さん:%s", time.Now().String()) // TODO: username送ってもらった方が良いかも
	if _, ok := r.subscribers[user]; ok {
		err := fmt.Errorf("`%s` has already been subscribed.", user)
		log.Print(err.Error())
		return nil, err
	}

	// チャンネルを作成し、リストに登録
	ch := make(chan *model.ChatMessage, 1)
	r.subscribers[user] = ch
	log.Printf("`%s` has been subscribed!", user)

	// コネクションが終了したら、このチャンネルを削除する
	go func() {
		<-ctx.Done()
		r.mutex.Lock()
		delete(r.subscribers, user)
		r.mutex.Unlock()
		log.Printf("`%s` has been unsubscribed.", user)
	}()

	return ch, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
