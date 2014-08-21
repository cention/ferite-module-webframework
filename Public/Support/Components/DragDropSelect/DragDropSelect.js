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
		return selected[0];
	};
	
	self.selectedItems = function() {
		var currentSelectedItem = Array();
		self.itemsEach(function(index, item) {
			if (item.hasClassName('selected')) {
				currentSelectedItem.push(item);
			}
		});
		return currentSelectedItem;
	}
	self.attachMouseDownActionWithValue = function( node, target, value ) {
		if( node ) {
			node.onclick = function( e ) {
				e = e || window.event;
				return GetComponent(target).action('click', e, value);
			};
		}
	};
	self.itemsEach( function( index, item ) {	
		self.attachMouseDownActionWithValue( item, self.identifier(), item);	  
	});
	var lastChecked = null;
	self.registerAction('click', function( event, item ) {		  
		var listItem = Array();
		var itemAsSelected = Array();
	        var currentSelectedItem = self.selectedItems();
		if (event.ctrlKey) {
			item.className = 'selected';
		} 
		else if (event.shiftKey) {
			listItem = self.getItemByCatagory(item);
			itemAsSelected = self.getValuesBetween(listItem,lastChecked,item);			
			self.clearSelected();
			for(i=0;i<itemAsSelected.length;++i)
				itemAsSelected[i].className ='selected';				
		}
		else {	    
			self.clearSelected();	    			
			item.className ='selected';
			lastChecked = item;
		}
	});
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
	self.getValuesBetween = function( array, a, b ) {
		var i1 = array.indexOf(a);
		var i2 = array.indexOf(b);
		var i;
		var r = [];

		if (i1 == -1 || i2 == -1) {
			return r;
		}
		if (i1 > i2) {
			var t = i2;
			i2 = i1;
			i1 = t;
		}
		for (i = i1; i <= i2; i++) {
			r.push(array[i]);
		}
		return r;
        };
	self.clearSelected = function() {
		var item = self.selectedItems();
		for(i=0;i<item.length;++i)
		item[i].className = '';				
	};
        self.getItemByCatagory = function( item ) {
		var list = Array();
		var listArray = Array();
		if (self.itemIsSelected(item))
			list = byId(self.identifier()).getElementsByTagName("li");
		else if (self.extraItemIsSelected(item))
			list = byId(self.identifier() + '_Target').getElementsByTagName("li");
		else 
			list = byId(self.identifier() + '_Source').getElementsByTagName("li");
		
		for(var i = 0; i < list.length; ++i)
			listArray.push(list[i]);
		
		return listArray; 
	};	
	
	self.updateSelected();
	
	return self;
}