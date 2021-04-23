package currentList

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type IntList struct {
	head *cList
	len  int32
}

func NewInt() *IntList {
	return &IntList{
		head: &cList{
			Value: 0,
		},
		len: 0,
	}
}

type cList struct {
	Value int
	sync.Mutex
	IsDelete bool
	Next     *cList
}

// 检查一个元素是否存在，如果存在则返回 true，否则返回 false
func (l *IntList) Contains(value int) bool {
	cur := l.head
	for {
		cur = (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cur.Next))))
		if cur == nil {
			break
		}
		if cur.Value > value {
			return false
		} else if cur.Value == value {
			return true
		}
	}
	return false
}

// 插入一个元素，如果此操作成功插入一个元素，则返回 true，否则返回 false
func (l *IntList) Insert(value int) bool {
	beforeCur := (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head))))
	cur := (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&beforeCur.Next))))
	for {
		for cur != nil {
			if value == cur.Value {
				return false
			} else if value < cur.Value {
				break
			} else {
				beforeCur = cur
				cur = (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cur.Next))))
			}
		}
		beforeCur.Lock()
		if (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&beforeCur.Next)))) != cur {
			cur = beforeCur
			beforeCur.Unlock()
			continue
		}
		break
	}
	node := &cList{
		Value: value,
		Next:  cur,
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&beforeCur.Next)), unsafe.Pointer(node))
	beforeCur.Unlock()
	atomic.AddInt32(&l.len, 1)
	return true
}

// 删除一个元素，如果此操作成功删除一个元素，则返回 true，否则返回 false
func (l *IntList) Delete(value int) bool {
	for {
		beforeCur := (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head))))
		cur := (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head.Next))))
		for {
			if cur == nil {
				return false
			}
			if value == cur.Value {
				break
			} else if value < cur.Value {
				return false
			} else {
				beforeCur = cur
				cur = (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&beforeCur.Next))))
			}
		}
		cur.Lock()
		if cur.IsDelete {
			cur.Unlock()
			return false
		}
		beforeCur.Lock()
		if (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&beforeCur)))).IsDelete ||
			(*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&beforeCur.Next)))) != cur {
			beforeCur.Unlock()
			cur.Unlock()
			continue
		}
		cur.IsDelete = true
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&beforeCur.Next)), unsafe.Pointer(cur.Next))
		atomic.AddInt32(&l.len, -1)
		beforeCur.Unlock()
		cur.Unlock()
		return true
	}
}

// 遍历此有序链表的所有元素，如果 f 返回 false，则停止遍历
func (l *IntList) Range(f func(value int) bool) {
	if l.Len() == 0 {
		return
	}
	cur := l.head
	for {
		cur = (*cList)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cur.Next))))
		if cur == nil {
			return
		}
		if !f(cur.Value) {
			return
		}
	}
}

// 返回有序链表的元素个数
func (l *IntList) Len() int {
	return (int)(atomic.LoadInt32(&l.len))
}
