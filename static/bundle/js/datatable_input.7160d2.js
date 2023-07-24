(function($){function calcDisableClasses(oSettings){var start=oSettings._iDisplayStart,length=oSettings._iDisplayLength,visibleRecords=oSettings.fnRecordsDisplay(),page=Math.ceil(start/length)||0,pages=Math.ceil(visibleRecords/length)||1,disableFirstPrevClass=page>0?"":oSettings.oClasses.sPageButtonDisabled,disableNextLastClass=page<pages-1?"":oSettings.oClasses.sPageButtonDisabled;return{first:disableFirstPrevClass,previous:disableFirstPrevClass,next:disableNextLastClass,last:disableNextLastClass}}function calcCurrentPage(oSettings){return Math.ceil(oSettings._iDisplayStart/oSettings._iDisplayLength)+1}function calcPages(oSettings){return Math.ceil(oSettings.fnRecordsDisplay()/oSettings._iDisplayLength)}var firstClassName="first",previousClassName="previous",nextClassName="next",lastClassName="last",paginateClassName="paginate",paginateInputClassName="paginate_input",paginateTotalClassName="paginate_total";$.fn.dataTableExt.oPagination.input={fnInit:function(oSettings,nPaging,fnCallbackDraw){var nWrap=document.createElement("ul"),nFirst=document.createElement("li"),nPrevious=document.createElement("li"),nNext=document.createElement("li"),nLast=document.createElement("li"),nInput=document.createElement("input"),nTotal=document.createElement("li"),nInfo=document.createElement("li"),language=oSettings.oLanguage.oPaginate,classes=oSettings.oClasses,info=language.info||"Page _INPUT_ of _TOTAL_";nFirst.innerHTML='<a tab-index="1" aria-controls="'+oSettings.sTableId+"_"+firstClassName+'" class="page-link">'+language.sFirst+"</a>",nPrevious.innerHTML='<a tab-index="1" aria-controls="'+oSettings.sTableId+"_"+previousClassName+'" class="page-link">'+language.sPrevious+"</a>",nNext.innerHTML='<a tab-index="1" aria-controls="'+oSettings.sTableId+"_"+nextClassName+'" class="page-link">'+language.sNext+"</a>",nLast.innerHTML='<a tab-index="1" aria-controls="'+oSettings.sTableId+"_"+lastClassName+'" class="page-link">'+language.sLast+"</a>",nWrap.className="pagination",nFirst.className=firstClassName+" "+classes.sPageButton,nPrevious.className=previousClassName+" "+classes.sPageButton,nNext.className=nextClassName+" "+classes.sPageButton,nLast.className=lastClassName+" "+classes.sPageButton,nInput.className=paginateInputClassName,nTotal.className=paginateTotalClassName,oSettings.sTableId!==""&&(nPaging.setAttribute("id",oSettings.sTableId+"_"+paginateClassName),nFirst.setAttribute("id",oSettings.sTableId+"_"+firstClassName),nPrevious.setAttribute("id",oSettings.sTableId+"_"+previousClassName),nNext.setAttribute("id",oSettings.sTableId+"_"+nextClassName),nLast.setAttribute("id",oSettings.sTableId+"_"+lastClassName)),nInput.type="text",info=info.replace(/_INPUT_/g,"</li>"+nInput.outerHTML+"<li>"),info=info.replace(/_TOTAL_/g,"</li>"+nTotal.outerHTML),nInfo.innerHTML=info,nWrap.appendChild(nFirst),nWrap.appendChild(nPrevious),$(nInfo).children().each(function(i,n){if(i===0){var nLi=document.createElement("li");nLi.className="paginate_input_wrap page-item",nLi.appendChild(n),nWrap.appendChild(nLi)}else i===1&&(n.className="paginate_of"),n.className=n.className+" page-item",nWrap.appendChild(n)}),nWrap.appendChild(nNext),nWrap.appendChild(nLast),nPaging.appendChild(nWrap),$(nFirst).click(function(){var iCurrentPage=calcCurrentPage(oSettings);iCurrentPage!==1&&(oSettings.oApi._fnPageChange(oSettings,"first"),fnCallbackDraw(oSettings))}),$(nPrevious).click(function(){var iCurrentPage=calcCurrentPage(oSettings);iCurrentPage!==1&&(oSettings.oApi._fnPageChange(oSettings,"previous"),fnCallbackDraw(oSettings))}),$(nNext).click(function(){var iCurrentPage=calcCurrentPage(oSettings);iCurrentPage!==calcPages(oSettings)&&(oSettings.oApi._fnPageChange(oSettings,"next"),fnCallbackDraw(oSettings))}),$(nLast).click(function(){var iCurrentPage=calcCurrentPage(oSettings);iCurrentPage!==calcPages(oSettings)&&(oSettings.oApi._fnPageChange(oSettings,"last"),fnCallbackDraw(oSettings))}),$(nPaging).find("."+paginateInputClassName).keyup(function(e){if(e.which===38||e.which===39?this.value++:(e.which===37||e.which===40)&&this.value>1&&this.value--,this.value===""||this.value.match(/[^0-9]/)!==null){this.value=this.value.replace(/[^\d]/g,"");return}var iNewStart=oSettings._iDisplayLength*(this.value-1);iNewStart<0&&(iNewStart=0),iNewStart>=oSettings.fnRecordsDisplay()&&(iNewStart=(Math.ceil(oSettings.fnRecordsDisplay()/oSettings._iDisplayLength)-1)*oSettings._iDisplayLength),oSettings._iDisplayStart=iNewStart,oSettings.oInstance.trigger("page.dt",oSettings),fnCallbackDraw(oSettings)}),$("span",nPaging).bind("mousedown",function(){return!1}),$("span",nPaging).bind("selectstart",function(){return!1});var iPages=calcPages(oSettings);iPages<=1&&$(nPaging).hide()},fnUpdate:function(oSettings){if(!oSettings.aanFeatures.p)return;var iPages=calcPages(oSettings),iCurrentPage=calcCurrentPage(oSettings),an=oSettings.aanFeatures.p;if(iPages<=1){$(an).hide();return}var disableClasses=calcDisableClasses(oSettings);$(an).show();var _newWidth="45px";$(an).find("."+paginateInputClassName).width(_newWidth),$(an).find("."+firstClassName).removeClass(oSettings.oClasses.sPageButtonDisabled).addClass(disableClasses[firstClassName]),$(an).find("."+previousClassName).removeClass(oSettings.oClasses.sPageButtonDisabled).addClass(disableClasses[previousClassName]),$(an).find("."+nextClassName).removeClass(oSettings.oClasses.sPageButtonDisabled).addClass(disableClasses[nextClassName]),$(an).find("."+lastClassName).removeClass(oSettings.oClasses.sPageButtonDisabled).addClass(disableClasses[lastClassName]),iPages===1&&($(an).find("."+nextClassName).removeClass(oSettings.oClasses.sPageButtonDisabled).addClass(disableClasses[nextClassName]),$(an).find("."+lastClassName).removeClass(oSettings.oClasses.sPageButtonDisabled).addClass(disableClasses[lastClassName])),$(an).find("."+paginateTotalClassName).html(iPages),$(an).find("."+paginateInputClassName).val(iCurrentPage)}}})(jQuery);
