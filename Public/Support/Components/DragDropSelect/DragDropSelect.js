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
	self.itemSelect = function( arrow ) {
		var item = self.selectedItems();
		for(i=0;i < item.length;i++ ) {
			if(item[i].parentNode.id == self.identifier() + '_Source' && arrow == '_FristRightArrow')
				self.node().appendChild(item[i]);
			if(item[i].parentNode.id == self.identifier() && arrow == '_SecondRightArrow') {
				if( byId(self.identifier() + '_Target') ) {
					byId(self.identifier() + '_Target').appendChild(item[i]);
				}
			}
		}
	};
	self.itemDeselect = function( arrow ) {
		var item = self.selectedItems();
		for(i=0;i<item.length;i++) {
			if(item[i].parentNode.id == self.identifier() && arrow == '_FristLeftArrow')
				byId(self.identifier() + '_Source').appendChild(item[i]);
			if(item[i].parentNode.id == self.identifier() + '_Target' && arrow == '_SeceondLeftArrow')
				self.node().appendChild(item[i]);
		}
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
		console.log(selected);
		return selected[0];
	};	
	self.selectedItems = function() {
		var currentSelectedItem = Array();
		self.itemsEach(function(index, item) {
			if (item.hasClassName('activated')) {
				currentSelectedItem.push(item);
			}
		});
		return currentSelectedItem;
	}
	
	byId(self.identifier() + '_FristLeftArrow').onclick = function() {
		self.itemDeselect('_FristLeftArrow');
		self.propagateChange();		
	};
	byId(self.identifier() + '_FristRightArrow').onclick = function() {
		self.itemSelect('_FristRightArrow');
		self.propagateChange();

	};
	if( byId(self.identifier() + '_SeceondLeftArrow') != null ) {	  
		byId(self.identifier() + '_SeceondLeftArrow').onclick = function() {
			self.itemDeselect('_SeceondLeftArrow');
			self.propagateChange();
		};
	}
	if( byId(self.identifier() + '_SecondRightArrow') != null ) {
		byId(self.identifier() + '_SecondRightArrow').onclick = function() {
			self.itemSelect('_SecondRightArrow');
			self.propagateChange();
		};
	}		
	self.updateSelected();
	
	return self;
}