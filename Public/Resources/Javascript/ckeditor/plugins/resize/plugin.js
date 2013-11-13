/*
 Copyright (c) 2003-2013, CKSource - Frederico Knabben. All rights reserved.
 For licensing, see LICENSE.html or http://ckeditor.com/license
*/
CKEDITOR.plugins.add("resize",{init:function(b){var f,g,n,o,a=b.config,q=b.ui.spaceId("resizer"),h=b.element?b.element.getDirection(1):"ltr";!a.resize_dir&&(a.resize_dir="vertical");void 0==a.resize_maxWidth&&(a.resize_maxWidth=3E3);void 0==a.resize_maxHeight&&(a.resize_maxHeight=3E3);void 0==a.resize_minWidth&&(a.resize_minWidth=750);void 0==a.resize_minHeight&&(a.resize_minHeight=250);if(!1!==a.resize_enabled){var c=null,i=("both"==a.resize_dir||"horizontal"==a.resize_dir)&&a.resize_minWidth!=a.resize_maxWidth,
l=("both"==a.resize_dir||"vertical"==a.resize_dir)&&a.resize_minHeight!=a.resize_maxHeight,j=function(d){var e=f,m=g,c=e+(d.data.$.screenX-n)*("rtl"==h?-1:1),d=m+(d.data.$.screenY-o);i&&(e=Math.max(a.resize_minWidth,Math.min(c,a.resize_maxWidth)));l&&(m=Math.max(a.resize_minHeight,Math.min(d,a.resize_maxHeight)));b.resize(i?e:null,m)},k=function(){CKEDITOR.document.removeListener("mousemove",j);CKEDITOR.document.removeListener("mouseup",k);b.document&&(b.document.removeListener("mousemove",j),b.document.removeListener("mouseup",
k))},p=CKEDITOR.tools.addFunction(function(d){c||(c=b.getResizable());f=c.$.offsetWidth||0;g=c.$.offsetHeight||0;n=d.screenX;o=d.screenY;a.resize_minWidth>f&&(a.resize_minWidth=f);a.resize_minHeight>g&&(a.resize_minHeight=g);CKEDITOR.document.on("mousemove",j);CKEDITOR.document.on("mouseup",k);b.document&&(b.document.on("mousemove",j),b.document.on("mouseup",k));d.preventDefault&&d.preventDefault()});
/*Mujibur: 2013-11-1 - Custom handler for mouse drag event to get the editor height to save to database.*/
var origin, startSize;
function dragHandler( evt )
{
    var dx = evt.data.$.screenX - origin.x;
    var dy = evt.data.$.screenY - origin.y;
    var internalWidth = startSize.width + dx * ( b.lang.dir == 'rtl' ? -1 : 1 );
    var internalHeight = startSize.height + dy;

    b.resize( Math.max( a.resize_minWidth, Math.min( internalWidth, a.resize_maxWidth ) ),
		    Math.max( a.resize_minHeight, Math.min( internalHeight, a.resize_maxHeight ) ) );
    
}

function dragEndHandler ( evt )
{
    CKEDITOR.document.removeListener( 'mousemove', dragHandler );
    CKEDITOR.document.removeListener( 'mouseup', dragEndHandler );
    
    if ( b.document )
    {
	b.document.removeListener( 'mousemove', dragHandler );
	b.document.removeListener( 'mouseup', dragEndHandler );
    }
    var dx = evt.data.$.screenX - origin.x;
    var dy = evt.data.$.screenY - origin.y;
    var internalWidth = startSize.width + dx * ( b.lang.dir == 'rtl' ? -1 : 1 );
    var internalHeight = startSize.height + dy;

    b.resize( Math.max( a.resize_minWidth, Math.min( internalWidth, a.resize_maxWidth ) ),
	      Math.max( a.resize_minHeight, Math.min( internalHeight, a.resize_maxHeight ) ) 
	    );
    
    mcam.fireCallbackRequest('RecordTextAreaSizes', function(value){}, {
	'question': '' + jQuery('#cke_QuestionWysiwyg iframe').contents().height(),
	'answer': '' + jQuery('#cke_AnswerWysiwyg iframe').contents().height()
    });
}
mujibur = CKEDITOR.tools.addFunction( function( $event ){ 
    //if ( !c )
    c = b.getResizable();
    startSize = { width : c.$.offsetWidth || 0, height : c.$.offsetHeight || 0 };
    origin = { x : $event.screenX, y : $event.screenY };

    CKEDITOR.document.on( 'mousemove', dragHandler );
    CKEDITOR.document.on( 'mouseup', dragEndHandler );

    if ( b.document )
    {
	    b.document.on( 'mousemove', dragHandler );
	    b.document.on( 'mouseup', dragEndHandler );
    }
}); 
b.on("destroy",function(){CKEDITOR.tools.removeFunction(p)});
b.on("uiSpace",function(a){if("bottom"==
a.data.space){var e="";i&&!l&&(e=" cke_resizer_horizontal");!i&&l&&(e=" cke_resizer_vertical");var c='<span id="'+q+'" class="cke_resizer'+e+" cke_resizer_"+h+'" title="'+CKEDITOR.tools.htmlEncode(b.lang.common.resize)+'" onmousedown="CKEDITOR.tools.callFunction('+mujibur+', event)">'+("ltr"==h?"◢":"◣")+"</span>";"ltr"==h&&"ltr"==e?a.data.html+=c:a.data.html=c+a.data.html}},b,null,100);b.on("maximize",function(a){b.ui.space("resizer")[a.data==CKEDITOR.TRISTATE_ON?"hide":"show"]()})}}});
