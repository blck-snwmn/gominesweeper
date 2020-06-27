package gominesweeper

import (
	"math/rand"
	"sync"
	"time"
)

type position struct {
	row, column int
}

func toCangeInfo(p position) ChangedInfo {
	return ChangedInfo{X: p.column, Y: p.row}
}

var noposition = position{-1, -1}

type cell struct {
	position      position
	to            map[position]chan<- ChangedInfo
	from          map[position]<-chan ChangedInfo
	pressed       bool
	pressedMux    sync.Mutex
	sendMux       sync.Mutex
	sendNum       int
	responseCh    chan ChangedInfo
	hasBomb       bool
	NearbyBombNum int
	notify        func(ChangedInfo)
}

// canPress return true if can press
// and set `c.pressed` tp true
func (c *cell) canPress() (pressed bool) {
	c.pressedMux.Lock()
	defer c.pressedMux.Unlock()
	pressed = c.pressed
	if pressed {
		return
	}
	c.pressed = true
	return
}

// press is called when cell pressed
// - if cell has bomb, call notify, return soon
// - if cell with a bomb nearby, call notify, return soon
// - if other cell, call send
func (c *cell) press(ignore position) ChangedInfo {
	if c.hasBomb {
		ci := toCangeInfo(c.position)
		ci.State = Bomb
		c.notify(ci)
		return ci
	}
	if c.NearbyBombNum > 0 {
		ci := toCangeInfo(c.position)
		ci.State = Opened
		ci.NumOfNearbyBomb = c.NearbyBombNum
		c.notify(ci)
		return ci
	}
	if c.canPress() {
		return ChangedInfo{X: -1, Y: -1}
	}
	ci := c.send(ignore)
	ci.NumOfNearbyBomb = c.NearbyBombNum
	c.notify(ci)
	return ci
}

// send send `ChangeInfo` except `ignore` and notify own
func (c *cell) send(ignore position) ChangedInfo {
	ci := ChangedInfo{X: c.position.column, Y: c.position.row}
	for p, ch := range c.to {
		if ignore == p {
			continue
		}
		ch <- ci
		close(ch)
	}
	// wait response
	for range c.responseCh {

	}
	// ci.NumOfNearbyBomb = c.NearbyBombNum
	// c.notify(ci)
	return ci
}

// responseRecievedMessage response recieved message
func (c *cell) responseRecievedMessage(recieved ChangedInfo) {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()
	c.sendNum--
	c.responseCh <- recieved
	// fmt.Printf("(%d, %d):recieve from(%d, %d)\n ", c.position.row, c.position.column, recieved.Y, recieved.X)
	if c.sendNum <= 0 {
		close(c.responseCh)
	}
}

// wake start goroutine each `cell.from`.
// when there goroutine recieve messsage, response.
func (c *cell) wake(h, w int) {
	for n, cf := range c.from {
		go func(hh, ww int, nn position, ccf <-chan ChangedInfo) {
			for recieved := range ccf {
				to := position{row: recieved.Y, column: recieved.X}
				if c.hasBomb {
					c.to[to] <- ChangedInfo{X: c.position.column, Y: c.position.row, State: Bomb}
					close(c.to[to])
					return
				}
				if c.NearbyBombNum > 0 {
					ci := ChangedInfo{X: c.position.column, Y: c.position.row, State: Opened, NumOfNearbyBomb: c.NearbyBombNum}
					c.to[to] <- ci
					close(c.to[to])
					c.canPress()
					c.notify(ci)
					return
				}
				// receive
				// 送信した数だけレスポンスが来るはず
				c.responseRecievedMessage(recieved)
				if c.canPress() {
					return
				}
				// response
				// すべてレスポンスされるまで待つ
				ci := c.send(to)
				ci.NumOfNearbyBomb = c.NearbyBombNum
				c.notify(ci)
				c.to[to] <- ci
				close(c.to[to])
				// fmt.Printf("recived (%d, %d):%v\n", hh, ww, nn)
				return
			}
		}(h, w, n, cf)
	}
}

type State int

const (
	NotOpen State = iota
	Opened
	Bomb
)

type ChangedInfo struct {
	X               int
	Y               int
	State           State
	NumOfNearbyBomb int
}
type Minesweeper struct {
	Height    int
	Width     int
	cells     [][]*cell
	sendCh    chan (<-chan ChangedInfo)
	recieveCh chan ChangedInfo
	buf       []ChangedInfo
	mux       sync.Mutex
}

func New(h, w, maxBombNum int) *Minesweeper {
	m := Minesweeper{
		Height:    h,
		Width:     w,
		buf:       []ChangedInfo{},
		sendCh:    make(chan (<-chan ChangedInfo)),
		recieveCh: make(chan ChangedInfo, h*w),
	}
	cells := make([][]*cell, m.Height)
	for i := 0; i < m.Height; i++ {
		cells[i] = make([]*cell, m.Width)
		for j := 0; j < m.Width; j++ {
			cells[i][j] = &cell{
				position: position{row: i, column: j},
				from:     map[position]<-chan ChangedInfo{},
				to:       map[position]chan<- ChangedInfo{},
				notify:   m.sendChange,
			}
		}
	}
	m.cells = cells
	m.registerAdjacentCell()
	m.wake()
	m.setBombs(maxBombNum)
	return &m
}

// GetNotify return channel that send `ChangeInfo`
func (m *Minesweeper) GetNotify() <-chan (<-chan ChangedInfo) {
	return m.sendCh
}

// PressCell called when cell pressed
func (m *Minesweeper) PressCell(row, column int) {
	c := m.cells[row][column]
	ch := make(chan ChangedInfo, len(c.to))
	go func() {
		c.press(noposition)
		m.getChangeCells(ch)
		// time.Sleep(100 * time.Millisecond)
	}()
	m.sendCh <- ch
}

func (m *Minesweeper) getChangeCells(ch chan<- ChangedInfo) {
	go func() {
		m.mux.Lock()
		defer m.mux.Unlock()
		defer close(ch)
		for _, ci := range m.buf {
			ch <- ci
		}
		m.buf = []ChangedInfo{}
	}()
}
func (m *Minesweeper) sendChange(ci ChangedInfo) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.buf = append(m.buf, ci)
}
func (m *Minesweeper) registerAdjacentCell() {
	f := func(src, dest *cell) {
		ch := make(chan ChangedInfo)
		src.to[dest.position] = ch
		dest.from[src.position] = ch
	}
	for h := 0; h < m.Height; h++ {
		for w := 0; w < m.Width; w++ {
			now := m.cells[h][w]
			if h > 0 && w > 0 {
				topLeft := m.cells[h-1][w-1]
				f(now, topLeft)
			}
			if h > 0 {
				top := m.cells[h-1][w]
				f(now, top)
			}
			if h > 0 && w < m.Width-1 {
				topRight := m.cells[h-1][w+1]
				f(now, topRight)
			}
			if w > 0 {
				left := m.cells[h][w-1]
				f(now, left)
			}
			if w < m.Width-1 {
				right := m.cells[h][w+1]
				f(now, right)
			}
			if h < m.Height-1 && w > 0 {
				btmLeft := m.cells[h+1][w-1]
				f(now, btmLeft)
			}
			if h < m.Height-1 {
				btm := m.cells[h+1][w]
				f(now, btm)
			}
			if h < m.Height-1 && w < m.Width-1 {
				btmRight := m.cells[h+1][w+1]
				f(now, btmRight)
			}
			now.responseCh = make(chan ChangedInfo, len(now.to))
			now.sendNum = len(now.to)
		}
	}
}

func (m *Minesweeper) setBombs(maxMombNum int) {
	rand.Seed(time.Now().UnixNano())
	bombNum := 0
	// 爆弾の数`bombNum`は`bombNum`<=`maxMombNum`
	// 同じcellへの更新が入ることを一旦許容
	for i := maxMombNum; i > 0; i-- {
		h := rand.Intn(m.Height)
		w := rand.Intn(m.Width)
		m.cells[h][w].hasBomb = true
		if h > 0 && w > 0 {
			topLeft := m.cells[h-1][w-1]
			topLeft.NearbyBombNum++
		}
		if h > 0 {
			top := m.cells[h-1][w]
			top.NearbyBombNum++
		}
		if h > 0 && w < m.Width-1 {
			topRight := m.cells[h-1][w+1]
			topRight.NearbyBombNum++
		}
		if w > 0 {
			left := m.cells[h][w-1]
			left.NearbyBombNum++
		}
		if w < m.Width-1 {
			right := m.cells[h][w+1]
			right.NearbyBombNum++
		}
		if h < m.Height-1 && w > 0 {
			btmLeft := m.cells[h+1][w-1]
			btmLeft.NearbyBombNum++
		}
		if h < m.Height-1 {
			btm := m.cells[h+1][w]
			btm.NearbyBombNum++
		}
		if h < m.Height-1 && w < m.Width-1 {
			btmRight := m.cells[h+1][w+1]
			btmRight.NearbyBombNum++
		}
		bombNum++
	}
}

func (m *Minesweeper) wake() {
	for h := 0; h < m.Height; h++ {
		for w := 0; w < m.Width; w++ {
			c := m.cells[h][w]
			c.wake(h, w)
		}
	}
}
