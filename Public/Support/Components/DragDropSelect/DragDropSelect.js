function ComponentDragDropSelect( id ) {
	var self = new _ComponentAbstractList(id);
	
	self._multiple = true;
	
	self.bind = function() { };
	self.items = function() {
		var items = [];
		var list = [];
		list = self.node().getElementsByTagName("li");
		for(var i = 0; i < list.length; ++i)
			items.push(list[i]);
		list = byId(self.identifier() + '_Source').getElementsByTagName("li");
		for(var i = 0; i < list.length; ++i)
			items.push(list[i]);
		if( byId(self.identifier() + '_Target') ) {
			list = byId(self.identifier() + '_Target').getElementsByTagName("li");
			for(var i = 0; i < list.length; ++i)
				items.push(list[i]);
		}
		return items;
	};
	self.itemIsSelected = function( item ) {
		return (item.parentNode.id == self.identifier() ? true : false);
	};
	self.extraItemIsSelected = function( item ) {
		return (item.parentNode.id == self.identifier() + '_Target' ? true : false);
	};
	self.itemValue = function( item ) {
		return item.getAttribute('value');
	};
	self.itemSelect = function( item, arrow ) {
		if(item.parentNode.id == self.identifier() + '_Source' && arrow == '_RightArrow')
			self.node().appendChild(item);
		if(item.parentNode.id == self.identifier() && arrow == '_RightArrow1') {
		        if( byId(self.identifier() + '_Target') ) {
				byId(self.identifier() + '_Target').appendChild(item);
			}
		}
	};
	self.itemDeselect = function( item, arrow ) {
		if(item.parentNode.id == self.identifier() && arrow == '_LeftArrow')
			byId(self.identifier() + '_Source').appendChild(item);
		if(item.parentNode.id == self.identifier() + '_Target' && arrow == '_LeftArrow1')
			self.node().appendChild(item);
	};
	self.updateFormValue = function() {	        
		if(node = byId('FormValue_' + self.identifier() + '_Order'))
			node.value = self.orderFormValue();
		if(node = byId('FormValue_' + self.identifier() + '_Selected'))
			node.value = self.selectedFormValue();
		if(node = byId('FormValue_' + self.identifier() + '_extraSelected'))
			node.value = self.extraSelectedFormValue();
	};
	self.submission = function() {
		var s = self.identifier() + '[order]=' + encodeURIComponent(self.orderFormValue()) +
			'&' + self.identifier() + '[selected]=' + encodeURIComponent(self.selectedFormValue())  +
			'&' + self.identifier() + '[extraSelected]=' + encodeURIComponent(self.extraSelectedFormValue());
		return s;
	};
	self.selectedItem = function() {
		var selected = self.getState('selected.list');
		return selected[0];
	};
	self.attachMouseDownActionWithValue = function( node, target, value ) {
		if( node ) {
			node.onmousedown = function( e ) {
				e = e || window.event;
				return GetComponent(target).action('click', e, value);
			};
		}
	};

	self._currentSelectedItem = null;
	self.itemsEach( function( index, item ) {
		self.attachMouseDownActionWithValue(item, self.identifier(), item);
	});
	self.registerAction('click', function( event, item ) {
		if( self._currentSelectedItem )
			self._currentSelectedItem.className = '';
		item.className = 'selected';
		self._currentSelectedItem = item;
	});

	byId(self.identifier() + '_LeftArrow').onclick = function() {
		if( self._currentSelectedItem ) {
			self.itemDeselect(self._currentSelectedItem, '_LeftArrow');
			self.propagateChange();
		}
	};
	byId(self.identifier() + '_RightArrow').onclick = function() {
		if( self._currentSelectedItem ) {
			self.itemSelect(self._currentSelectedItem, '_RightArrow');
			self.propagateChange();
		}
	};
	if( byId(self.identifier() + '_LeftArrow1') != null ) {	  
		byId(self.identifier() + '_LeftArrow1').onclick = function() {
			if( self._currentSelectedItem ) {
				self.itemDeselect(self._currentSelectedItem, '_LeftArrow1');
				self.propagateChange();
			}
		};
	}
	if( byId(self.identifier() + '_RightArrow1') != null ) {
		byId(self.identifier() + '_RightArrow1').onclick = function() {
			if( self._currentSelectedItem ) {
				self.itemSelect(self._currentSelectedItem, '_RightArrow1');
				self.propagateChange();
			}
		};
	}
	self.updateSelected();
	
	return self;
}