/*
MIT License
Copyright(c) 2022 Futurewei Cloud

	Permission is hereby granted,
	free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
	including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
	to whom the Software is furnished to do so, subject to the following conditions:
	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
package datastore

import (
	"context"

	"github.com/go-redis/redis/v9"
	//"github.com/futurewei-cloud/merak/services/merak-compute/interfaces"
)

type Store struct {
	Client redis.Client
}

func (store *Store) Get(ctx context.Context, id string, field string) (string, error) {
	res := store.Client.HGet(ctx, id, field)
	return res.Val(), nil
}
func (store *Store) Update(ctx context.Context, id string, obj []byte) error {
	store.Client.HSet(ctx, id, obj)
	return nil
}
func (store *Store) Delete(ctx context.Context, id string) error {
	store.Client.HDel(ctx, id)
	return nil
}

func (store *Store) GetList(ctx context.Context, id string) ([]byte, error) {
	return nil, nil
}
func (store *Store) AddToList(ctx context.Context, id string, json []byte) error {
	return nil
}
func (store *Store) DeleteList(ctx context.Context, id string) error {
	return nil
}

func (store *Store) GetSet(ctx context.Context, id string) ([]byte, error) {
	return nil, nil
}
func (store *Store) AddToSet(ctx context.Context, id string, json []byte) error {
	return nil
}
func (store *Store) DeleteSet(ctx context.Context, id string) error {
	return nil
}
