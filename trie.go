package jrr

import "strings"

// 前缀树实现路由匹配

type node struct {
	pattern  string  // 待匹配路由，例如 /p/:lang
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 是否精确匹配，part 含有 : 或 * 时为true
}

// 与普通的树不同，为了实现动态路由匹配，加上了isWild这个参数。
// 当匹配 /p/go/doc/这个路由时，
// 第一层节点，p精准匹配到了p
// 第二层节点，go模糊匹配到:lang，那么将会把lang这个参数赋值为go，
// 继续下一层匹配。

// matchChild 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// matchChildren 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// insert 将新注册的路由增加到前缀树上
// 插入功能很简单，递归查找每一层的节点，如果没有匹配到当前part的节点，则新建一个
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) <= height {
		n.pattern = pattern
		return
	}
	// 当前需要匹配的part
	curPart := parts[height]

	// 将当前part(curPart)与当前结点(n)下所有子结点的part进行匹配，看是否匹配成功
	child := n.matchChild(curPart)
	if child == nil { // 没有匹配到，为当前节点(n)新增加一个子节点
		child = &node{part: curPart, isWild: curPart[0] == '*' || curPart[0] == ':'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// search 找到最深处满足条件的子节点
func (n *node) search(parts []string, height int) *node {
	if len(parts) <= height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	// 从当前结点下所有子节点中查找
	part := parts[height]
	children := n.matchChildren(part) // 从子节点中找出所有符合part的子节点
	for _, child := range children {
		ret := child.search(parts, height+1)
		if ret != nil {
			return ret
		}
	}
	return nil
}
