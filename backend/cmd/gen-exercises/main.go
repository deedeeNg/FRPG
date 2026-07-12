// Command gen-exercises generates French exercises and writes them to stdout as
// JSON Lines (one exercise per line) — the exercises.jsonl handoff format. It is
// the v0 stand-in for the Python generation pipeline; the import tool
// (cmd/import-exercises) consumes the same format.
//
//	go run ./cmd/gen-exercises            # 51 items (uniform across skill points)
//	go run ./cmd/gen-exercises -n 90 > exercises.jsonl
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"frpg-backend/internal/generate"
)

func main() {
	n := flag.Int("n", 0, "total exercises to generate (0 = 15 per skill pack, spread uniformly)")
	levels := flag.String("levels", "content/levels/a1.yaml", "path to the level pack YAML")
	flag.Parse()

	pack, err := generate.LoadPack(*levels)
	if err != nil {
		log.Fatalf("load level pack: %v", err)
	}
	total := *n
	if total <= 0 {
		total = 15 * len(pack.Teaches)
	}
	exs := generate.Generate(pack, total)

	enc := json.NewEncoder(os.Stdout)
	dist := map[string]int{}
	for _, e := range exs {
		if err := enc.Encode(e); err != nil {
			log.Fatalf("encode %s: %v", e.ID, err)
		}
		dist[e.Contrast.SkillPoint]++
	}

	// Report the skill distribution on stderr so it doesn't pollute the jsonl.
	fmt.Fprintf(os.Stderr, "generated %d exercises\n", len(exs))
	for _, sp := range sortedKeys(dist) {
		fmt.Fprintf(os.Stderr, "  %-16s %d\n", sp, dist[sp])
	}
}

func sortedKeys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
