// script.aculo.us multi drag and drop add on
//
// See http://www.github.com/panter/scriptaculous-multidrag/
// and http://script.aculo.us/
// for more information
//
// script.aculo.us is freely distributable under the terms of an MIT-style
// license. For details, see the web site:http://script.aculo.us/
//
// The multi d&d add on code is in the public domain.
//
// This add-on enables multiple selections for scripty draggables and
// sortables.
// 
// Please note that the sortable add on needs a slightly patched version of
// dragdrop.js.

//////////// utility methods ////////////////////////////////////////////////

// returns an array of elements that are in state activated.
function getActivatedElementIds() {
  activated = [].concat($$('.activated'));
  activated = activated.pluck('id').uniq();
  return activated;
}
 
// add needed listeners for multidrag on the given element
function prepareMultidrag(element) {
  element.observe('mousedown', function(e) {
    if (e.shiftKey &&
        !element.hasClassName('activated') &&
        $$('.activated').size() > 0) {
      activateSiblings(element.previousSiblings()) ||
        activateSiblings(element.nextSiblings());
    }
    element.toggleClassName('activated');
  });
}
 
// activates the elements of the given siblings array till the first activated
// element is found (needed for shift click)
function activateSiblings(siblings) {
  if (!siblings.find(function(s) { return s.hasClassName('activated') })) {
    return false;
  }
  siblings.each(function(s) {
      if (s.hasClassName('activated')) { throw $break; }
      s.addClassName('activated');
    });
  return true;
}
 
function deactivateAll() {
  $$('.activated').each(function(e) { e.removeClassName('activated'); });
}
 
function activateAll() {
  Draggables.drags.each(function(d) { d.element.addClassName('activated'); });
}

// the effect to apply when drag revert happens: this places to element
// immediatly on its targeted position
function immediateRevertEffect(element, top_offset, left_offset) {
  postop = parseFloat(element.style.top) - top_offset;
  posleft = parseFloat(element.style.left) - left_offset;
  element.setStyle({
    left: posleft.round() + 'px',
    top:  postop.round()  + 'px'
  });
}
 
//////////// observers //////////////////////////////////////////////////////

// scripty draggable observer that paints a box on the current draggable,
// indicating how many objects are activated currently
var MultidragObserver = Class.create({
    initialize: function(container) {
    // reset all droppables on container click
	Event.observe(container, 'click', function(e) {
		if (container.id == e.element().id) {
			deactivateAll();
		}
		$$('.activated').each(function(e) {
			if (container.id != e.parentNode.id) {
				deactivateAll();
			}
		});
		if (e.shiftKey || e.ctrlKey && !e.element().hasClassName('activated') && $$('.activated').size() > 0) {
			activateSiblings(e.element().previousSiblings()) || 
			activateSiblings(e.element().nextSiblings());
		} 
		else deactivateAll(); 
	        if(e.element().id != container.id) {
			e.element().addClassName('activated');
		}
	});
	// add mouse down listener to activate draggables
	//Draggables.drags.each(function(draggable) {
	  //prepareMultidrag(draggable.element);
	//});
    },
 
  // called on drag start: displays a info box on the draggable showing how many
  // elements are activated for this drag if this is a multi element drag
  onStart: function(eventName, draggable, domEvent) {
    //draggable.element.addClassName('activated');
   // draggable._clone.addClassName('activated');
    $$('.dragcount').each(function(e) { e.remove() });
    activated = getActivatedElementIds();
    if (activated.length > 1) {
      info = new Element('div', { 'class': 'dragcount' });
      info.insert(activated.length);
      draggable.element.insertBefore(info, draggable.element.firstChild);
    }
  },
 
  // called on drag end: removes drag count info boxes and reorders if the drag
  // target was the sortable.
  onEnd: function(eventName, draggable, domEvent) {
    $$('.dragcount').each(function(e) { e.remove() });
  }
 
});
 
 
// scrypty draggable observer that supports reordering more than one element
var MultisortObserver = Class.create({
  initialize: function(container) {
    this.container = container;
  },
 
  // called on drag start: remembers element order 
  onStart: function(eventName, draggable, domEvent) {
    this.lastSequence = Sortable.sequence(this.container);
  },
 
  // called on drag end: reorders the sortable elements when multiple elements
  // were dropped.
  onEnd: function(eventName, draggable, domEvent) {
    // do nothing if the drop area has reveived the elements
    //if (null == Sortable._marker) { return; }
    draggableId = draggable.element.id.substring(1+draggable.element.id.lastIndexOf('_'));
    origindex = this.lastSequence.indexOf(draggableId);
    // do nothing if the draggable is still on its place
    newSequence = Sortable.sequence(this.container);
    if (origindex == newSequence.indexOf(draggableId)) {
        return;
    }
    // drop the other activated elements near the just dropped draggable
    rightSibling = draggable.element;
    parentNode = draggable.element.parentNode;
    $$('.activated').each(function(e) {
      if (draggable.element.id != e.id) {
        id = e.id.substring(1+e.id.indexOf('_'));
        if (this.lastSequence.indexOf(id) < origindex) {
          parentNode.insertBefore(e, draggable.element);
        } else {
          parentNode.insertBefore(e, rightSibling.nextSibling);
          rightSibling = e;
        }
      }
    }, this);
  }
 
});
