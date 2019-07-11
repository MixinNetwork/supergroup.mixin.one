package plugin

import (
	"sort"
	"sync"
)

// group := models.Shortcut.CreateGroup("group1", "Group 1", "组1", 0)
// group.CreateItem("item1", "Item 1", "项1", "https://...", 0)
// fmt.Println(models.Shortcut.AllGroups())

type shortcut struct{}

var Shortcut shortcut

type ShortcutItem struct {
	ID       string `json:"id"`
	LabelEn  string `json:"label_en"`
	LabelZh  string `json:"label_zh"`
	URL      string `json:"url"`
	Sequence int    `json:"-"`
}

type ShortcutGroup struct {
	ID         string          `json:"id"`
	LabelEn    string          `json:"label_en"`
	LabelZh    string          `json:"label_zh"`
	Items      []*ShortcutItem `json:"items"`
	Sequence   int             `json:"-"`
	itemsIndex map[string]*ShortcutItem
}

var (
	shortcutGroups []*ShortcutGroup = []*ShortcutGroup{
		&ShortcutGroup{
			ID:         "default",
			LabelEn:    "Default",
			LabelZh:    "群操作",
			Items:      []*ShortcutItem{},
			itemsIndex: map[string]*ShortcutItem{},
		},
	}
	shortcutGroupsIndex map[string]*ShortcutGroup = map[string]*ShortcutGroup{"default": shortcutGroups[0]}
	shortcutRWMutex     sync.RWMutex
)

func (shortcut) AllGroups() []*ShortcutGroup {
	return shortcutGroups
}

func (shortcut) FindGroup(id string) *ShortcutGroup {
	shortcutRWMutex.RLock()
	defer shortcutRWMutex.RUnlock()

	return shortcutGroupsIndex[id]
}

func (shortcut) CreateGroup(id, labelEn, labelZh string, sequence int) *ShortcutGroup {
	if g := Shortcut.FindGroup(id); g != nil {
		return g
	}

	shortcutRWMutex.Lock()
	defer shortcutRWMutex.Unlock()

	group := &ShortcutGroup{
		ID:       id,
		LabelEn:  labelEn,
		LabelZh:  labelZh,
		Items:    []*ShortcutItem{},
		Sequence: sequence,
	}

	shortcutGroups = append(shortcutGroups, group)
	sort.SliceStable(shortcutGroups, func(i, j int) bool {
		return shortcutGroups[i].Sequence < shortcutGroups[j].Sequence
	})
	Shortcut.reindex()
	return group
}

func (shortcut) reindex() {
	shortcutGroupsIndex = map[string]*ShortcutGroup{}
	for _, g := range shortcutGroups {
		shortcutGroupsIndex[g.ID] = g
	}
}

func (g *ShortcutGroup) CreateItem(id, labelEn, labelZh, url string, sequence int) *ShortcutItem {
	if s := g.FindItem(id); s != nil {
		return s
	}

	shortcutRWMutex.Lock()
	defer shortcutRWMutex.Unlock()

	item := &ShortcutItem{
		ID:       id,
		LabelEn:  labelEn,
		LabelZh:  labelZh,
		URL:      url,
		Sequence: sequence,
	}

	g.Items = append(g.Items, item)
	sort.SliceStable(g.Items, func(i, j int) bool {
		return g.Items[i].Sequence < g.Items[j].Sequence
	})
	g.reindex()
	return item
}

func (g *ShortcutGroup) FindItem(id string) *ShortcutItem {
	shortcutRWMutex.RLock()
	defer shortcutRWMutex.RUnlock()

	return g.itemsIndex[id]
}

func (g *ShortcutGroup) reindex() {
	g.itemsIndex = map[string]*ShortcutItem{}
	for _, item := range g.Items {
		g.itemsIndex[item.ID] = item
	}
}
