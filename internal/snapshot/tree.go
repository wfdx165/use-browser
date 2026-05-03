package snapshot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chromedp/chromedp"
)

type Node struct {
	Ref      string   `json:"ref,omitempty"`
	Role     string   `json:"role"`
	Name     string   `json:"name,omitempty"`
	Value    string   `json:"value,omitempty"`
	Children []string `json:"children"`
}

type Snapshot struct {
	Root  *Node
	Nodes map[string]*Node
	Refs  map[string]int64
}

func BuildTree(ctx context.Context, interactiveOnly, compact bool, depthLimit int) (*Snapshot, error) {
	var rawJSON string
	if err := chromedp.Run(ctx, chromedp.Evaluate(buildJS(interactiveOnly), &rawJSON)); err != nil {
		return nil, err
	}

	var result struct {
		Root  string `json:"root"`
		Nodes []Node `json:"nodes"`
	}

	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("snapshot parse: %w", err)
	}

	snap := &Snapshot{
		Nodes: make(map[string]*Node),
		Refs:  make(map[string]int64),
	}

	for i := range result.Nodes {
		n := &result.Nodes[i]
		snap.Nodes[n.Ref] = n
		if n.Ref == result.Root {
			snap.Root = n
		}
	}

	return snap, nil
}

func buildJS(interactive bool) string {
	i := "false"
	if interactive {
		i = "true"
	}
	return `(function(i){var items=[],counter=0;function getRole(e){if(e.getAttribute('role'))return e.getAttribute('role');var t=e.tagName.toLowerCase();if(t==='input'){var ty=e.getAttribute('type')||'text';if(ty==='checkbox')return'checkbox';if(ty==='radio')return'radio';if(ty==='submit'||ty==='button'||ty==='reset')return'button';return'textbox';}var m={button:'button',a:'link',input:'textbox',textarea:'textarea',select:'combobox',option:'option',h1:'heading',h2:'heading',h3:'heading',h4:'heading',h5:'heading',h6:'heading',img:'image',form:'form',ul:'list',ol:'list',li:'listitem',nav:'navigation',header:'banner',footer:'contentinfo',main:'Main',section:'region',article:'Article'};return m[t]||'generic';}var intRoles={button:1,link:1,textbox:1,combobox:1,listbox:1,radio:1,checkbox:1,switch:1,slider:1,menuitem:1,tab:1,treeitem:1,option:1,searchbox:1,textarea:1,gridcell:1,row:1,columnheader:1,rowheader:1,heading:1,list:1,listitem:1};function getName(e){var n=e.getAttribute('aria-label')||e.getAttribute('placeholder')||'';if(n)return n.trim();if(e.children.length===0){var t=e.textContent.trim();if(t.length>0&&t.length<100)return t;}return'';}function walk(e){if(!e||!e.tagName)return null;var r=getRole(e);if(r===''||r==='none')return null;if(i&&!intRoles[r]&&r!=='link'&&r!=='heading'&&r!=='generic'&&r!=='banner'&&r!=='navigation'&&r!=='Region'&&r!=='Main'&&r!=='Article'&&r!=='contentinfo'&&r!=='complementary'&&r!=='form')return null;counter++;var ref='@e'+counter;var ch=[];for(var j=0;j<e.children.length;j++){var c=walk(e.children[j]);if(c)ch.push(c.ref);}var n={ref:ref,role:r,name:getName(e),value:e.value||'',children:ch};return n;}try{var root=walk(document.documentElement);if(!root){return JSON.stringify({root:'',nodes:[]});}items.unshift(root);return JSON.stringify({root:root.ref,nodes:items});}catch(e){return JSON.stringify({error:e.message});}})("` + i + `")`
}
