(function() {
  var interactiveOnly = false;
  var compactMode = false;

  var interactiveRoles = {
    'button': true, 'link': true, 'textbox': true, 'combobox': true,
    'listbox': true, 'radio': true, 'checkbox': true, 'switch': true,
    'slider': true, 'menuitem': true, 'tab': true, 'treeitem': true,
    'option': true, 'searchbox': true, 'textarea': true,
    'gridcell': true, 'row': true, 'columnheader': true, 'rowheader': true,
    'Main': true, 'heading': true, 'list': true, 'listitem': true,
    'navigation': true, 'banner': true, 'contentinfo': true, 'region': true,
    'Article': true, 'complementary': true
  };

  var tagToRole = {
    'button': 'button', 'a': 'link',
    'input': 'textbox', 'textarea': 'textarea', 'select': 'combobox',
    'option': 'option',
    'h1': 'heading', 'h2': 'heading', 'h3': 'heading',
    'h4': 'heading', 'h5': 'heading', 'h6': 'heading',
    'img': 'image', 'form': 'form', 'ul': 'list', 'ol': 'list',
    'li': 'listitem', 'nav': 'navigation', 'header': 'banner',
    'footer': 'contentinfo', 'main': 'Main', 'section': 'region',
    'article': 'Article', 'aside': 'complementary',
    'table': 'table', 'th': 'columnheader', 'td': 'gridcell',
    'tr': 'row', 'label': 'label',
    'html': 'RootWebArea', 'body': 'WebArea',
    'div': 'generic', 'span': 'generic', 'p': 'generic',
    'pre': 'generic', 'code': 'generic', 'br': 'generic'
  };

  var counter = 0;
  var nodes = [];

  function getRole(el) {
    var tag = el.tagName ? el.tagName.toLowerCase() : '';
    var role = el.getAttribute('role');
    if (role) return role;
    if (tag === 'input') {
      var t = el.getAttribute('type');
      if (t === 'checkbox') return 'checkbox';
      if (t === 'radio') return 'radio';
      if (t === 'submit' || t === 'button' || t === 'reset') return 'button';
      if (t === 'range') return 'slider';
      if (t === 'email' || t === 'search' || t === 'password' || t === 'tel' || t === 'url' || t === 'number') return 'textbox';
      return 'textbox';
    }
    return tagToRole[tag] || tag || 'generic';
  }

  function getName(el) {
    var name = el.getAttribute('aria-label');
    if (name && name.trim()) return name.trim();
    var labId = el.getAttribute('aria-labelledby');
    if (labId) {
      var lab = document.getElementById(labId);
      if (lab) return lab.textContent.trim();
    }
    if (!el.children || el.children.length === 0) {
      var text = el.textContent;
      if (text && text.trim().length > 0 && text.trim().length < 100) return text.trim();
    }
    var placeholder = el.getAttribute('placeholder');
    if (placeholder) return placeholder;
    return '';
  }

  function getValue(el) {
    var tag = el.tagName ? el.tagName.toLowerCase() : '';
    if (tag === 'input' || tag === 'textarea') {
      return el.value || '';
    }
    return '';
  }

  var minVisibleSize = 1;

  function walk(el, depth) {
    if (!el || !el.tagName || !el.nodeType || el.nodeType !== 1) return null;

    var role = getRole(el);
    if (role === 'none') return null;

    if (interactiveOnly && depth > 0 && !interactiveRoles[role] && role !== 'heading') {
      var childResults = [];
      var childElements = el.children;
      for (var i = 0; i < childElements.length; i++) {
        var cr = walk(childElements[i], depth + 1);
        if (cr) { childResults.push(cr); nodes.push(cr); }
      }
      if (childResults.length === 0) return null;
      var children = [];
      for (var j = 0; j < childResults.length; j++) children.push(childResults[j].ref);
      counter++;
      var wrapperRef = '@e' + counter;
      return { ref: wrapperRef, role: role, name: '', value: '', children: children, props: {} };
    }

    counter++;
    var ref = '@e' + counter;

    var children = [];
    var childElements = el.children;
    if (childElements) {
      for (var k = 0; k < childElements.length; k++) {
        var childResult = walk(childElements[k], depth + 1);
        if (childResult) {
          children.push(childResult.ref);
          nodes.push(childResult);
        }
      }
    }

    return {
      ref: ref, role: role, name: getName(el), value: getValue(el),
      children: children, props: {}
    };
  }

  var treeRoot;
  try {
    treeRoot = walk(document.documentElement, 0);
  } catch(e) {
    return JSON.stringify({error: e.message, root: '', nodes: []});
  }

  if (!treeRoot) {
    return JSON.stringify({root: '', nodes: []});
  }

  nodes.unshift(treeRoot);
  return JSON.stringify({ root: treeRoot.ref, nodes: nodes });
})()
