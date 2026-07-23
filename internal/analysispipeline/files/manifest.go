package files

import (
	"sort"
	"strings"
)

func BuildManifests(paths []LogicalPath) []StructureManifest {
	byRoot := map[string][]string{}
	var topLevel []string

	for _, p := range paths {
		if p.ArchiveRoot == "" {
			topLevel = append(topLevel, p.RelativePath)
			continue
		}
		byRoot[p.ArchiveRoot] = append(byRoot[p.ArchiveRoot], p.RelativePath)
	}

	var manifests []StructureManifest
	for root, relPaths := range byRoot {
		sort.Strings(relPaths)
		manifests = append(manifests, StructureManifest{
			ArchiveRoot: root,
			Tree:        renderTree(root, relPaths),
		})
	}
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].ArchiveRoot < manifests[j].ArchiveRoot
	})

	if len(topLevel) > 0 {
		sort.Strings(topLevel)
		var b strings.Builder
		b.WriteString("## Submission files\n")
		for _, name := range topLevel {
			b.WriteString("- ")
			b.WriteString(name)
			b.WriteByte('\n')
		}
		manifests = append([]StructureManifest{{
			Tree: b.String(),
		}}, manifests...)
	}

	return manifests
}

type treeNode struct {
	children map[string]*treeNode
	isFile   bool
}

func renderTree(root string, relPaths []string) string {
	rootNode := &treeNode{children: map[string]*treeNode{}}
	for _, rel := range relPaths {
		parts := strings.Split(strings.Trim(rel, "/"), "/")
		current := rootNode
		for i, part := range parts {
			if part == "" {
				continue
			}
			if current.children[part] == nil {
				current.children[part] = &treeNode{children: map[string]*treeNode{}}
			}
			current = current.children[part]
			if i == len(parts)-1 {
				current.isFile = true
			}
		}
	}

	var b strings.Builder
	b.WriteString("## Project structure (")
	b.WriteString(root)
	b.WriteString(")\n")
	b.WriteString(root)
	b.WriteString("/\n")
	writeTree(&b, rootNode, "  ")
	return b.String()
}

func writeTree(b *strings.Builder, n *treeNode, indent string) {
	names := sortedKeys(n.children)
	for _, name := range names {
		child := n.children[name]
		if child.isFile && len(child.children) == 0 {
			b.WriteString(indent)
			b.WriteString(name)
			b.WriteByte('\n')
			continue
		}
		b.WriteString(indent)
		b.WriteString(name)
		b.WriteString("/\n")
		writeTree(b, child, indent+"  ")
	}
}

func sortedKeys(m map[string]*treeNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func collectPaths(inline []InlineSection, attachments []Attachment) []LogicalPath {
	seen := map[string]struct{}{}
	var paths []LogicalPath
	add := func(p LogicalPath) {
		key := p.String()
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		paths = append(paths, p)
	}
	for _, section := range inline {
		add(section.Path)
	}
	for _, file := range attachments {
		add(file.Path)
	}
	return paths
}
