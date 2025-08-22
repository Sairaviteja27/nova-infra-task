package wallet

import (
	"context"
	"sync"
	"time"

	"github.com/sairaviteja27/nova-infra-task/types"
	"github.com/sairaviteja27/nova-infra-task/utils"
	"golang.org/x/sync/singleflight"
)

type FetchFunc func(ctx context.Context, addr string) (types.Result, error)

type Service struct {
	cache *utils.Cache
	sf    singleflight.Group
	fetch FetchFunc
}

func NewService(ttl time.Duration, fetch FetchFunc) *Service {
	return &Service{
		cache: utils.NewCache(ttl),
		fetch: fetch,
	}
}

func (s *Service) Get(ctx context.Context, addr string) (types.Result, error) {
	if v, ok := s.cache.Get(addr); ok {
		return v, nil
	}
	res, err, _ := s.sf.Do(addr, func() (interface{}, error) {
		if v, ok := s.cache.Get(addr); ok {
			return v, nil
		}
		v, err := s.fetch(ctx, addr)
		if err == nil {
			s.cache.Set(addr, v)
		}
		return v, err
	})
	if err != nil {
		return types.Result{}, err
	}
	return res.(types.Result), nil
}

func (s *Service) FetchMany(ctx context.Context, addresses []string) ([]types.Result, map[string]error) {
	type kv struct {
		addr string
		val  types.Result
		err  error
	}

	uniq := make(map[string]struct{}, len(addresses))
	for _, a := range addresses {
		uniq[a] = struct{}{}
	}

	out := make(chan kv, len(uniq))
	wg := sync.WaitGroup{}

	for addr := range uniq {
		a := addr
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := s.Get(ctx, a)
			out <- kv{addr: a, val: v, err: err}
		}()
	}

	wg.Wait()
	close(out)

	results := make([]types.Result, 0, len(uniq))
	errs := make(map[string]error)

	for item := range out {
		if item.err != nil {
			errs[item.addr] = item.err
			continue
		}
		results = append(results, item.val)
	}
	return results, errs
}
