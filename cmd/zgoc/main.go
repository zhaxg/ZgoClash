// Command zgoc is a client of MyGo's automation plugin: it lists the app's
// windows, reads the elements of their native UI, acts on them, and captures
// them as PNG images.
//
//	go run ./cmd/zgoc tree
//	go run ./cmd/zgoc press core-toggle
//	go run ./cmd/zgoc shot dashboard.png
package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type request struct {
	ID     any            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type response struct {
	ID     any             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type node struct {
	ID      string   `json:"id"`
	Parent  string   `json:"parent,omitempty"`
	Role    string   `json:"role"`
	Key     any      `json:"key,omitempty"`
	Label   string   `json:"label,omitempty"`
	Value   string   `json:"value,omitempty"`
	Bounds  bounds   `json:"bounds"`
	States  []string `json:"states,omitempty"`
	Actions []string `json:"actions,omitempty"`
}

type bounds struct {
	X, Y, Width, Height float64
}

type window struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Kind          string `json:"kind"`
	Focused       bool   `json:"focused"`
	Visible       bool   `json:"visible"`
	Bounds        bounds `json:"bounds"`
	ContentBounds bounds `json:"contentBounds"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: zgoc <command> [args]")
		os.Exit(2)
	}
	sock := os.Getenv("MYGO_AUTOMATION")
	if sock == "" {
		sock = "C:\\Users\\ZHAXG\\AppData\\Local\\Temp\\zgoc.sock"
	}
	c, err := dial(sock)
	if err != nil {
		fail("connect %s: %v", sock, err)
	}
	defer c.c.Close()

	switch os.Args[1] {
	case "windows":
		res := call(c, "windows", nil)
		printJSON(res)
	case "tree":
		res := call(c, "tree", nil)
		if res == nil {
			return
		}
		var t struct {
			Nodes []node `json:"nodes"`
		}
		if err := json.Unmarshal(res, &t); err != nil {
			fail("tree: %v", err)
		}
		printTree(t.Nodes)
	case "press":
		if len(os.Args) < 3 {
			fail("press needs a key")
		}
		res := call(c, "act", map[string]any{"key": os.Args[2], "action": "press"})
		printJSON(res)
	case "click":
		// click <x> <y> in the window's content, in DIPs.
		if len(os.Args) < 4 {
			fail("click needs x and y")
		}
		var x, y float64
		fmt.Sscanf(os.Args[2], "%f", &x)
		fmt.Sscanf(os.Args[3], "%f", &y)
		res := call(c, "input", map[string]any{
			"events": []map[string]any{{"type": "click", "x": x, "y": y}},
		})
		printJSON(res)
	case "shot":
		if len(os.Args) < 3 {
			fail("shot needs a file name")
		}
		res := call(c, "capture", nil)
		if res == nil {
			return
		}
		var r struct {
			PNG    string `json:"png"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		}
		if err := json.Unmarshal(res, &r); err != nil {
			fail("capture: %v", err)
		}
		data, err := base64.StdEncoding.DecodeString(r.PNG)
		if err != nil {
			fail("decode png: %v", err)
		}
		if err := os.WriteFile(os.Args[2], data, 0o644); err != nil {
			fail("write: %v", err)
		}
		fmt.Printf("saved %s (%dx%d)\n", os.Args[2], r.Width, r.Height)
	case "idle":
		res := call(c, "idle", nil)
		printJSON(res)
	default:
		fail("unknown command %q", os.Args[1])
	}
}

func dial(path string) (*conn, error) {
	nc, err := net.DialTimeout("unix", path, 5*time.Second)
	if err != nil {
		return nil, err
	}
	return &conn{c: nc, r: bufio.NewReader(nc)}, nil
}

type conn struct {
	c net.Conn
	r *bufio.Reader
}

var nextID = 0

// call sends a request and waits for its answer.
func call(c *conn, method string, params map[string]any) json.RawMessage {
	nextID++
	req := request{ID: nextID, Method: method, Params: params}
	line, err := json.Marshal(req)
	if err != nil {
		fail("marshal: %v", err)
	}
	if _, err := c.c.Write(append(line, '\n')); err != nil {
		fail("write: %v", err)
	}
	for {
		raw, err := c.r.ReadBytes('\n')
		if err != nil {
			fail("read: %v", err)
		}
		var resp response
		if err := json.Unmarshal(raw, &resp); err != nil {
			continue
		}
		if resp.ID == nil || fmt.Sprint(resp.ID) != fmt.Sprint(nextID) {
			continue
		}
		if resp.Error != nil {
			fail("%s: %s: %s", method, resp.Error.Code, resp.Error.Message)
		}
		return resp.Result
	}
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fail("marshal: %v", err)
	}
	fmt.Println(string(data))
}

// printTree shows the elements of the window, with the keys a test can use.
// The tree is a flat list with parent references, so it is rebuilt here.
func printTree(nodes []node) {
	children := map[string][]node{}
	var roots []node
	byID := map[string]node{}
	for _, n := range nodes {
		byID[n.ID] = n
	}
	for _, n := range nodes {
		if n.Parent == "" {
			roots = append(roots, n)
			continue
		}
		children[n.Parent] = append(children[n.Parent], n)
	}
	for _, n := range roots {
		printNode(n, 0, children)
	}
}

func printNode(n node, depth int, children map[string][]node) {
	indent := strings.Repeat("  ", depth)
	label := n.Label
	if n.Value != "" && n.Value != n.Label {
		label += " = " + n.Value
	}
	key := ""
	if n.Key != nil {
		key = fmt.Sprintf(" key=%v", n.Key)
	}
	states := ""
	if len(n.States) > 0 {
		states = " [" + strings.Join(n.States, ",") + "]"
	}
	fmt.Printf("%s%s%s \"%s\" (%.0f,%.0f %.0fx%.0f)%s\n",
		indent, n.Role, key, label, n.Bounds.X, n.Bounds.Y, n.Bounds.Width, n.Bounds.Height, states)
	for _, c := range children[n.ID] {
		printNode(c, depth+1, children)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
