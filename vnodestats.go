package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx"
)

// Int64Slice attaches the methods of sort.Interface to []int64, sorting in
// increasing order.

type Int64Slice []int64

func (s Int64Slice) Len() int           { return len(s) }
func (s Int64Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s Int64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Sort is a convenience method.
func (s Int64Slice) Sort() {
	sort.Sort(s)
}

type Stats struct {
	vmap    map[int64]int
	count   int64
	avgDist float64
}

func (s *Stats) consumeVNode(rows *pgx.Rows) error {
	var vnode int64
	err := rows.Scan(&vnode)
	if err != nil {
		return err
	}

	s.vmap[vnode]++
	s.count++

	return nil
}

func (s *Stats) printStats() {
	fmt.Fprintf(os.Stderr, "Total number of vnodes: %d\n", s.count)
	fmt.Fprintf(os.Stderr, "Number of unique vnodes sampled: %d\n", len(s.vmap))
	fmt.Fprintf(os.Stderr, "Average Distance Between vnodes is %f\n", s.avgDist)
}

var conn *pgx.Conn

func main() {
	conn, err := pgx.Connect(extractConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
		os.Exit(1)
	}

	tx, err := conn.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { tx.Rollback() }()
	_, err = tx.Exec("DECLARE vnode_window NO SCROLL CURSOR FOR SELECT _vnode FROM manta WHERE _vnode IS NOT NULL;")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { tx.Exec("CLOSE vnode_window;") }()

	s := &Stats{
		vmap: make(map[int64]int),
	}

	start := time.Now()
	for {
		rows, err := tx.Query("FETCH 10000 FROM vnode_window;")
		if err != nil {
			log.Fatal(err)
		}

		// Calling rows.Next() is the only way to check if there are more rows
		// returned from a query, so we have to check if there are no rows to break
		// from our otherwise infinite loop
		if !rows.Next() {
			break
		} else {
			s.consumeVNode(rows)
		}
		for rows.Next() {
			s.consumeVNode(rows)
		}
	}
	queryTime := time.Now()

	var keys Int64Slice
	for k := range s.vmap {
		keys = append(keys, k)
	}
	keys.Sort()

	var lastVNode int64
	for n, key := range keys {
		// If this is the first vnode, don't average
		if n == 0 {
			continue
		}
		dist := key - lastVNode
		s.avgDist = math.Abs(float64(dist)+s.avgDist) / 2
		lastVNode = key
	}
	readTime := time.Now()

	s.printStats()
	fmt.Fprintf(os.Stderr, "Query Performed in %s\n", queryTime.Sub(start).String())
	fmt.Fprintf(os.Stderr, "Read Performed in %s\n", readTime.Sub(queryTime).String())
}

func extractConfig() pgx.ConnConfig {
	var config pgx.ConnConfig

	config.Host = os.Getenv("DB_HOST")
	if config.Host == "" {
		config.Host = "localhost"
	}

	config.User = os.Getenv("DB_USER")
	if config.User == "" {
		config.User = "postgres"
	}

	config.Password = os.Getenv("DB_PASSWORD")
	if config.Database == "" {
		config.Database = "postgres"
	}

	config.Database = os.Getenv("DB_DATABASE")
	if config.Database == "" {
		config.Database = "moray"
	}

	return config
}
