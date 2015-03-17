
var helperCurrent = '';

function ComponentHelper( id, target, alternativeTarget ) {
	var self = new Component(id);
	var popup = $(id + '.popup');
	var visible = false;
	
	self.highlight = null;

	var parent = self.node().parentNode;
	
	for( ; parent && parent.tagName.toLowerCase() != 'tr'; parent = parent.parentNode )
	 	;
	if( parent ) {
		self.highlight = parent;
	}
	
	self.show = function() {
		if( id != helperCurrent ) {
			if( helperCurrent ) {
				_(helperCurrent).hide();
			}
			Element.clonePosition(popup, id, {
					setWidth: false,
					setHeight: false,
					offsetLeft: self.node().offsetWidth + 10,
					offsetTop: 0 - self.node().offsetHeight + document.viewport.getScrollOffsets().top
				});
			Element.show(popup);

			if( self.highlight )
				self.highlight.style.backgroundColor = '#e1ffe4';

			$A([ target, alternativeTarget ]).each(function( t ) {
				if( _(t ) ) {
					_(t).focus();
				} else if( $(t) ) {
					try {
						$(t).focus();
					} catch( e ) {
					}
				}
			});

			helperCurrent = id;
			visible = true;
		}
	};
	
	self.hide = function() {
		helperCurrent = '';
		if( self.highlight )
			self.highlight.style.backgroundColor = '#FFF';
		$A([ target, alternativeTarget ]).each(function( t ) {
			if( _(t) ) {
				_(t).blur();
			} else if( $(t) ) {
				$(t).blur();
			}
		});
		Element.hide(popup);
		visible = false;
	};

	self.setHighlight = function( node ) {
		self.highlight = node;
	};

	self.registerAction('click', function(event) {
		if( visible ) {
			self.hide();
		} else {
			self.show();
		}
	});
	$A([ target, alternativeTarget ]).each(function( t ) {
		if( _(t) ) {
			_(t).registerAction('focus', function() {
				self.show();
			});
			_(t).registerAction('blur', function() {
				self.hide();
			});
		} else if( $(t) ) {
			$(t).onfocus = function(event) {
				self.show();
			};
			$(t).onblur = function(event) {
				self.hide();
			};
		}
	});

	return self;
}
