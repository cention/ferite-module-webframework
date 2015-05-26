function ComponentCkeditor( id ) {
	var self = new ComponentTextfield(id);
	
	self.setDefaultState('text-value');
	self.setState('text-value', '');
	
	self.editor = CKEDITOR.instances[id];
	self.editorIsReady = false;
	self.editorToolbar = 'OneRow';
	self.haveUnsetData = false;
	self.hideWhenReady = false;
	
	self.config = {
		skin: 'kama',
		enterMode: CKEDITOR.ENTER_DIV,
		shiftEnterMode: CKEDITOR.ENTER_DIV,
		forcePasteAsPlainText: false,
		allowedContent: true,
		disableNativeSpellChecker: false,
		browserContextMenuOnCtrl: true,
		colorButton_enableMore: false,
		resize_dir: 'both',
		/*resize_enabled: true,*/
		fontSize_sizes:
			'8/8pt;' +
			'9/9pt;' +
			'10/10pt;' +
			'11/11pt;' +
			'12/12pt;' +
			'14/14pt;' +
			'16/16pt;' +
			'18/18pt;' +
			'20/20pt;' +
			'22/22pt;' +
			'24/24pt;' +
			'26/26pt;' +
			'28/28pt;' +
			'36/36pt;' +
			'48/48pt',
		font_names:
			'Arial/Arial, Helvetica, sans-serif;' +
			'Arial black;' +
			'Calibri/Calibri, sans-serif;' +
			'Comic Sans MS/Comic Sans MS, cursive;' +
			'Courier New/Courier New, Courier, monospace;' +
			'Georgia/Georgia, serif;' +
			'Lucida Console/Lucida Console, monospace;' +
			'Lucida Sans Unicode/Lucida Sans Unicode, Lucida Grande, sans-serif;' +
			'Tahoma/Tahoma, Geneva, sans-serif;' +
			'Times New Roman/Times New Roman, Times, serif;' +
			'Trebuchet MS/Trebuchet MS, Helvetica, sans-serif;' +
			'Verdana/Verdana, Geneva, sans-serif',
		toolbar_Empty: [],
		toolbar_SingleRow: [
			[ 'SpellCheckerLanguage' ],
			[ 'SpellChecker' ]
		],
		toolbar_OneRow: [
			[ 'Bold', 'Italic', 'Underline', 'Strike' ],
			[ 'NumberedList', 'BulletedList' ],
			[ 'Indent', 'Outdent', 'Blockquote' ],
			[ 'JustifyLeft', 'JustifyCenter', 'JustifyRight' ],
			[ 'Link', 'Image', 'Youtube', 'HorizontalRule' ], [ 'PasteFromWord' ],
			[ 'Font' ], [ 'FontSize' ],
			[ 'TextColor', 'BGColor' ],
			[ 'SpellCheckerLanguage' ], [ 'SpellChecker' ], [ 'Resize' ]
		],
		toolbar_TwoRow: [
			[ 'Bold', 'Italic', 'Underline', 'Strike' ],
			[ 'NumberedList', 'BulletedList' ],
			[ 'Indent', 'Outdent', 'Blockquote' ],
			[ 'JustifyLeft', 'JustifyCenter', 'JustifyRight' ], [ 'PasteFromWord' ],
			[ 'Link', 'Image', 'Youtube', 'HorizontalRule' ],
			[ 'Font' ], [ 'FontSize' ],'/',
			[ 'TextColor', 'BGColor' ],
			[ 'SpellCheckerLanguage' ], [ 'SpellChecker' ], ['Resize']
		],
		toolbar: 'OneRow',
		plugins:
			'basicstyles,' +
			'blockquote,' +
			'clipboard,' +
			'colorbutton,' +
			'contextmenu,' +
			'enterkey,' +
			'font,' +
			'horizontalrule,' +
			'htmlwriter,' +
			'indent,' +
			'justify,' +
			'link,' +
			'list,' +
			'magicline,' +
			'pastetext,' +
			/* 'removeformat,' + */
			'showborders,' +
			'tab,' +
			/*'table,' + */
			/*'tabletools,' + */
			'toolbar,' +
			'undo,' +
			'wysiwygarea,' +
			'cention_spellcheck,' +
			'cention_image,' +
			'pastefromword,' + 
			'resize,' +
			'youtube',
		contentsCss: [
			WFServerURI + 'Resources/CSS/jquery.spellchecker.css',
			WFServerURI + 'Resources/Javascript/ckeditor/contents.css',
			WFServerURI + 'Support/Components/Ckeditor/Ckeditor.css',
		],
		spellCheckLanguages: []
	};

	self.setLanguage = function( language ) {
		self.config.language = language;
	}

	self.setFontSize = function( size ){
		if( size ){
			self.config.fontSize_defaultLabel = '' + size;
			CKEDITOR.addCss('.cke_editable { font-size: '+ size +'; }');
		}
	};
	
	self.setFontFamily = function( family ){
		if( family ){
			self.config.font_defaultLabel = '' + family;
			CKEDITOR.addCss('.cke_editable { font-family: '+ family +'; }');
		}
	}

	self.setLanguages = function( list ) {
	    self.config.spellCheckLanguages = list;
	};
	self.setImages = function( list ) {
		if( self.editor ) {
			self.editor.___fileArchiveImages = list;
		}
	};
	self.setShowResize = function( resize ) {
		if( resize == true ) {
			self.config.resize_enabled = true;
		} else {
			self.config.resize_enabled = false;
		}
	};
	self.setItemHeight = function( height ){
		self.config.height = height;
	};
	self.showBasicToolbar = function() { };
	self.showAdvancedToolbar = function() { };
	self.setTwoRowToolbar = function( value ) {
		if( value ) {
			self.editorToolbar = 'TwoRow';
			self.config.toolbar = self.editorToolbar;
		}
	};
	self.setShowToolbar = function( value ) {
		if( !value ) {
			self.editorToolbar = 'Empty';
			self.config.toolbar = self.editorToolbar;
		}
	};
	self.setSimpleToolbar = function( value ) {
		if( !value ) {
			self.editorToolbar = 'SingleRow';
			self.config.toolbar = self.editorToolbar;
		}
	};

	self.bind = function() {
		self.attachFocusAction();
		self.attachBlurAction();
	};

	self.attachFocusAction = function() {
		if( self.editor ) {
			self.editor.on('focus', function() {
				self.action('focus');
			});
		}
	};
	self.attachBlurAction = function() {
		if( self.editor ) {
			self.editor.on('blur', function() {
				self.action('blur');
			});
		}
	};

	self.setReadOnly = function( value ) {};
	self.setEnabled = function( value ) {};

	self.focus = function() {
		if( self.editor ) {
			self.editor.focus();
			return true;
		}
		return false;
	};
	self.blur = function() {
		return false;
	};
	self.show = function() {
		var node = $('cke_' + self.identifier());
		if( node ) {
			node.show();
		}
	};
	self.hide = function() {
		if( self.editorIsReady ) {
			var node = $('cke_' + self.identifier());
			if( node ) {
				node.hide();
			}
		} else {
			self.hideWhenReady = true;
		}
	};
	
	self.empty = function() {
		if( self.editor ) {
			if( self.editor.getData() )
				return false;
		}
		return true;
	};
	
	var previousGetState = self.getState;
	self.getState = function( state ) {
		if( state == 'text-value' ) {
			return self.editor.getData();
		}
		return previousGetState(state);
	};
	
	var previousSetState = self.setState;
	self.setState = function( state, value ) {
		if( state == 'text-value' ) {
			var dataSet = false;
			self._states['text-value'] = value;
			if( self.editor ) {
				if( self.editorIsReady ) {
					self.editor.setData(value);
					self.editor.updateElement();
					dataSet = true;
				} 
			}
			if( !dataSet ) {
				self.haveUnsetData = true;
			}
		} else {
			previousSetState(state, value);
		}
	};
	
	self.updateVisual = function() {};
	self.updateFormValue = function() {
		if( self.editor ) {
			self.editor.updateElement();
		}
	};
	
	var previousActivate = self.activate;
	self.activate = function() {
		self.editor = CKEDITOR.replace(self.identifier(), self.config);

		// Tobias 2015-03-17: The bind() function is called by SetComponent() however at the time
		//                    This happens the editor isn't created which means no event handlers
		//                    can be attached to it. The bind() function is therefore called one
		//                    more time here.
		self.bind();

		self.editor.on('contentDom', function() {
			self.editorIsReady = true;
			if( self.haveUnsetData ) {
				self.editor.setData(self._states['text-value']);
				self.editor.updateElement();
				self.haveUnsetData = false;
			}
			if( self.hideWhenReady ) {
				$('cke_' + self.identifier()).hide();
				self.hideWhenReady = false;
			}
		});
				
		self.editor.on('contentDomUnload', function() {
			self.editorIsReady = false;
		});
		previousActivate();
	};
	
	return self;
}

CKEDITOR.addCss('body { margin-left: 5px; margin-right: 5px; margin-top: 3px; margin-bottom: 3px; }');
CKEDITOR.addCss('p { margin-top: 0; }');
