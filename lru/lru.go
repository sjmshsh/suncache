package lru

import "container/list"

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// 双向链表节点的数据类型，在链表中仍然保存每一个值对应的key
// 好处在于，淘汰队首节点的时候，需要用key从字典中删除对应的映射
type entry struct {
	key   string
	value Value
}

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes int64                         // 允许使用的最大的内存
	nbytes   int64                         // 当前已经使用的内存
	ll       *list.List                    // Go标准库中实现的双向链表
	cache    map[string]*list.Element      // 键是字符串，值是双向链表中对应节点的指针
	onEvited func(key string, value Value) // 某条记录被移除时的回调函数
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		onEvited: onEvicted,
	}
}

// Get look ups a key's value
// 如果键对应的链表还存在，则将对应节点移动到队尾，并返回查找到的值
// 这里约定front为队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest item
// c.ll.Back() 取到队首节点，从链表中删除。
// delete(c.cache, kv.key)，从字典中 c.cache 删除该节点的映射关系。
// 更新当前所用的内存 c.nbytes。
// 如果回调函数 OnEvicted 不为 nil，则调用回调函数。
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvited != nil {
			c.onEvited(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	// 如果键存在，则更新对应节点的值，并将该节点移到队尾。
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 不存在则是新增场景，首先队尾添加新节点 &entry{key, value},
		// 并字典中添加 key 和节点的映射关系。
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 更新 c.nbytes，如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点。
	if c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}
