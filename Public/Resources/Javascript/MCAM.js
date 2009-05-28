
function MCAMOutputSystem() {
	this.messageBox = function( message, extra ) {
		alert(message);
	};
	this.errorBox = function( message, extra ) {
		alert(message);
	};
}

function MCAM() { // Multiple Channel AJAX Mechanism
	this.requester = null;
	this.handlers = new Array();
	this.dirtyList = new Array();
	this._dirtyList = new Array();
	this.callbacks = new Array();
	this.url = '';
	this.loading = 0;
	
	this.createRequestObject = function() {
		try { 
			return new XMLHttpRequest(); 
		} catch (error) { 
			try { 
				return new ActiveXObject("Microsoft.XMLHTTP"); 
			} catch (error) {
				return null;
			}
		}
		return null;
	};
	this.registerType = function( handlerType, handler ) {
		this.handlers[handlerType] = handler;
	};
	this.registerCallback = function( requestName, handler ) {
		this.callbacks[requestName] = handler;
	};

	this.lastChannel = '';
	this.outputSystem = new MCAMOutputSystem();
	this.setOutput = function( os ) {
		this.outputSystem = os;
	};
	this.handleChannel = function( requester, node ) {
		var id = '';
		var type = '';
		var content = '';
		var i = 0;

		for( i = 0; i < node.childNodes.length; i++ ) {
			switch( node.childNodes[i].tagName ) {
				case 'id':
					id = node.childNodes[i].firstChild.nodeValue;
					break;
				case 'type':
					type = node.childNodes[i].firstChild.nodeValue;
					break;
				case 'content':
					content = node.childNodes[i];
					break;
			}
		}

		var rval = this.handlers[type]( requester, id, type, content );
		this.lastChannel = 'Channel: (type:' + type + ',id:' + id + ')';
		return rval;
	};
	this.handleEvent = function( requester ) {
		if( !requester.abortedByUser ) {
			switch( requester.readyState ) {
				case 4: {
					if( requester.status == 200 ) {
						var i = 0, lastChannel = 0;
						var successful = true;
//						try {
							var root = requester.responseXML.firstChild;
							if( requester.responseXML.documentElement )
								root = requester.responseXML.documentElement;
							if( root ) {
								for( i = 0; i < root.childNodes.length; i++ ) {
									lastChannel = i;
									if( root.childNodes[i].tagName == 'channel' && !this.handleChannel( requester, root.childNodes[i] ) && successful ) {
										successful = false;
										break;
									}
								}
							}
//						} catch ( e ) {
//							this.outputSystem.errorBox( 'Error Decoding MCAM Packet: (channel #' + lastChannel + ')\n' + e.message + '\n',  requester.responseText );
//						}
						if( !successful ) {
							this.outputSystem.errorBox( 'Error Decoding MCAM Packet.', requester.responseText );
						}
					} else if( requester.status != 404 && requester.status > 0 ) {
						this.outputSystem.errorBox('All going wrong -> ' + requester.status + ' : ' + requester.mcamURL, '');
					}
					this.dirtyList = new Array();
					this._dirtyList = new Array();
					this.toggleLoading(false);
					break;
				}
			}
		}
	};
	this.abort = function( requester ) {
		this.toggleLoading(false);
		requester.abortedByUser = true;
		requester.abort();
	};
	this.registerDirtyComponent = function ( id ) {
		this.dirtyList.push(id);
	};
	this.setComponentIsDirty = function( object ) {
		this._dirtyList.push( object );
	};
	this.toggleLoading = function( onoff) {
		var node = document.getElementById('mcam_status');
		
		if( onoff ) {
			this.loading++;
		} else {
			this.loading--;
		}
		
		if( this.loading < 0 )
			this.loading = 0;
		
		if( node ) {
			if( this.loading ) {
				if( node.builtByMCAM ) {
					node.style.display = 'block';
					node.style.top = '5px';
					node.style.left = '5px';
				} else {
					node.style.display = '';
				}
			}
			else
				node.style.display = 'none';
		}
	}
	this.fireBackgroundEvent = function( component, event_type, extra ) {
		var url = this.getTargetURL();
		var self = this;
		var parameters = '';
		var i = 0;
		
		/* Old lists */
		for( i = 0; i < this.dirtyList.length; i++ ) {
			var nodeid = this.dirtyList[i];
			var node = document.getElementById(nodeid);
			if( node && IsValidFormComponent(node) ) {
				parameters += SerializeFormComponent( nodeid, node );
			}
		}
		/* New list */
		for( i = 0; i < this._dirtyList.length; i++ ) {
			parameters += '&' + this._dirtyList[i].submission();
			this._dirtyList[i].setClean();
		}
		this.toggleLoading(true);
		
		var requester = this.createRequestObject();
		requester.open( "POST", url ); 
		requester.setRequestHeader( 'Content-Type','application/x-www-form-urlencoded' );
		requester.onreadystatechange = function() { 
			self.handleEvent(requester);
		};
		requester.send(  'uieventcomponent='+ component +
								'&uieventdata=' + event_type +
								'&uieventextra=' + encodeURIComponent(extra) +
								parameters );
	};
	this.fireForegroundEvent = function( component, event_type, extra ) {
		submitComponentForm( component, event_type, extra );
	};
	this.setTargetURL = function( url ) {
		this.url = url;
	};
	this.getTargetURL = function() {
		var url = this.url;
		if( url == '' ) {
			url = window.location.href;
			if( url.indexOf('/-/') > 0 ) {
				url = url.substring(0, url.indexOf('-/'));
			}
		}
		return url;
	};
	this.fireReplaceRequest = function( request, target, new_parameters ) {
		return this.fireReplaceRequestWithCallback( request, null, target, new_parameters );
	};
	this.fireCallbackRequest = function( request, callback, new_parameters ) {
		return this.fireReplaceRequestWithCallback( request, callback, '', new_parameters );
	};
	this.createProgressDiv = function( node, label ) {
		var pos = findPos(node);
		var div = document.createElement('div');
		div.innerHTML = '<img src="' + uriForServerImageResource('loading_animation_liferay.gif') + '" />';
		div.id = label + '_status';
		div.style.display = 'block';
		div.style.position = 'absolute';
		div.style.zIndex = 100;
		div.style.left = (pos[0] + node.offsetWidth - 70) + 'px';
		div.style.top = pos[1] + 'px';
		return div;
	};
	this.fireReplaceRequestWithCallback = function( request, callback, target, new_parameters ) {
		var url = this.getTargetURL() + '/-/MCAM/' + request;
		var self = this;
		var parameters = '';
		var i = 0;

		/* Old lists */
		for( i = 0; i < this.dirtyList.length; i++ ) {
			var nodeid = this.dirtyList[i];
			var node = document.getElementById(nodeid);
			if( node && IsValidFormComponent(node) ) {
				parameters += SerializeFormComponent( nodeid, node );
			}
		}
		/* New list */
		for( i = 0; i < this._dirtyList.length; i++ ) {
			parameters += '&' + this._dirtyList[i].submission();
			this._dirtyList[i].setClean();
		}

		for( key in new_parameters ) {
			parameters += '&' + key + "=" + encodeURIComponent(new_parameters[key]);
		}
		
		this.toggleLoading(true);
		
		var requester = this.createRequestObject();
		requester.mcamURL = url;
		requester.open( "POST", url ); 
		requester.setRequestHeader( 'Content-Type','application/x-www-form-urlencoded' );
		requester.onreadystatechange = function() { 
			self.handleEvent(requester); 
		};

		var status_div;
		if( target != '' ) {
			var target_node = document.getElementById(target);
			status_div = this.createProgressDiv( target_node, target );
			wfinsertAdjacentElement( target_node, "afterEnd", status_div );
		}
		requester.mcamCallback = function( id, type, content ) {
			if( callback )
				callback( (content.firstChild ? content.firstChild.nodeValue : '') );
			if( target ) {
				status_div.parentNode.removeChild(status_div);
				return self.handlers['SetContent']( null, target, '', content );
			}
			return true;
		};
		
		requester.send( parameters );
		
		return requester;
	};
	this.componentRequest = function( c, r ) {
		return '' + c + '.' + r;
	};
	/*** SETUP ***/
	var self = this;
	this.registerType( 'Result', function( requester, id, type, content ) {
		return requester.mcamCallback( id, type, content );
	});
	this.registerType( 'Replace', function( requester, id, type, content ) { 
		var node = document.getElementById(id);
		if( node ) {
			if( content.firstChild )
				wfinsertAdjacentHTML( node, 'replace', content.firstChild.nodeValue );
			else
				node.parentNode.removeChild( node );
			return true;
		} 
		return true;
	} );
	this.registerType( 'SetContent', function( requester, id, type, content ) {
		var node = document.getElementById(id);
		if( node ) {
			node.innerHTML = '';
			if( content.firstChild )
				wfinsertAdjacentHTML( node, 'beforeEnd', content.firstChild.nodeValue );
			return true;
		}
		return false;
	});
	this.registerType( 'SetValue', function( requester, id, type, content ) {
		var node = document.getElementById(id);
		if( node ) {
			document.getElementById(id).value = content.firstChild.nodeValue;
			return true;
		}
		return false;
	});
	this.registerType( 'Script', function( requester, id, type, content ) {
		try {
			eval( content.firstChild.nodeValue );
			return true;
		} catch( e ) {
			return false;
		}
	});
	this.registerType( 'Error', function( requester, id, type, content ) {
		var errorMessage = content.firstChild.nodeValue;
		alert( 'MCAM.Error: ' + errorMessage );
		return true;
	});
	this.log = function( value ) {
		if( $('log') ) {
			$('log').innerHTML += value + '<br />';
		}
		try {
			if( console ) {
				console.log(value);
			}
		} catch(e) {}
	};
	this.logError = function( message ) {
		try {
			mcam.log(message);
			var previous = mcam.url;
			mcam.url = urlForApplicationAction('ReportedErrors');
			var requestor = mcam.fireCallbackRequest(
				'ReportJavascriptError', 
				function( value ) { return true; }, 
				{ 'backtrace': message, 'action': window.location.href }
			);
			mcam.url = previous;
			return requestor;
		} catch(e) {
			return null;
		}
	};
};

var mcam = new MCAM();

var loadFunction = function() {
	if( !document.getElementById('mcam_status') ) {
		var s = BrowserWindowSize();
		var n = document.createElement('div');
		n.id = 'mcam_status';
		n.style.top = '5px';
		n.style.left = '5px';
		n.style.display = 'none';
		n.style.width = '5px';
		n.style.backgroundColor = '#FFF';
		n.style.border = '0px solid #FFF';
		n.style.padding = '0px';
		n.style.color = '#33F';
		n.style.position = 'absolute';
		n.style.zIndex = 1000;
		n.builtByMCAM = true;
		n.innerHTML = '<img style="margin:0px;" src="' + uriForServerImageResource('loading_animation_liferay.gif') + '" />';
		
		var formElements = document.getElementsByName('uicomponentform');
		wfinsertAdjacentElement( formElements[0], "afterEnd", n ); 
	}
};
registerLoadFunction(loadFunction);
