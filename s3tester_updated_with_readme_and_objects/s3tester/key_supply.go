package main

import (
    "bufio"
    "context"
    "errors"
    "math/rand"
    "os"
    "sync"
    "time"
)

type KeySupply struct {
    keys []string
    mu   sync.Mutex
    idx  int64
    rand *rand.Rand
}

func NewKeySupplyFromFile(path string) (*KeySupply, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    var keys []string
    for scanner.Scan() {
        line := scanner.Text()
        if line == "" {
            continue
        }
        keys = append(keys, line)
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    if len(keys) == 0 {
        return nil, errors.New("no keys loaded from file")
    }
    return &KeySupply{
        keys: keys,
        rand: rand.New(rand.NewSource(time.Now().UnixNano())),
    }, nil
}

func (ks *KeySupply) Next(ctx context.Context, mode string, threadID int64) (string, error) {
    ks.mu.Lock()
    defer ks.mu.Unlock()
    if len(ks.keys) == 0 {
        return "", errors.New("no keys available")
    }
    switch mode {
    case "roundrobin":
        k := ks.keys[ks.idx%int64(len(ks.keys))]
        ks.idx++
        return k, nil
    case "random":
        return ks.keys[ks.rand.Intn(len(ks.keys))], nil
    default:
        k := ks.keys[ks.idx%int64(len(ks.keys))]
        ks.idx++
        return k, nil
    }
}

func (ks *KeySupply) Size() int { return len(ks.keys) }
