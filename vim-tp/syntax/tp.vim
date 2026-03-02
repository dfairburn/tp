" syntax/tp.vim - Syntax highlighting for tp template files
" Provides YAML-like highlighting for template structure, embedded JSON
" highlighting in the body section, embedded GraphQL highlighting in the
" query section, and Go template {{ ... }} highlighting throughout.

if exists('b:current_syntax')
  finish
endif

" --- Include JSON syntax for body embedding ---
syn include @jsonSyntax syntax/json.vim
unlet! b:current_syntax

" --- Include GraphQL syntax for query embedding ---
syn include @graphqlSyntax syntax/graphql.vim
unlet! b:current_syntax

" --- Go Template Syntax (global, highest priority) ---
" Defined early so they can appear in any context. The containedin=ALL
" ensures they override JSON, GraphQL, header values, URL values, etc.
syn region tpTemplateExpr matchgroup=tpTemplateBraces start='{{' end='}}' containedin=ALL contains=tpTemplateKeyword,tpTemplateDot,tpTemplateString,tpTemplatePipe,tpTemplateFunc

" Fix: The built-in jsonString and jsonKeyword regions only contain
" jsonEscape, so tpTemplateExpr cannot appear inside them despite
" containedin=ALL. We clear and redefine them to include tpTemplateExpr.
syn clear jsonStringMatch
syn clear jsonString
syn clear jsonKeywordMatch
syn clear jsonKeyword
syn match jsonStringMatch /"\([^"]\|\\\"\)\+"\ze[[:blank:]\r\n]*[,}\]]/ contains=jsonString,tpTemplateExpr
syn region jsonString oneline matchgroup=jsonQuote start=/"/  skip=/\\\\\|\\"/  end=/"/ contains=jsonEscape,tpTemplateExpr contained
syn match  jsonKeywordMatch /"\([^"]\|\\\"\)\+"[[:blank:]\r\n]*\:/ contains=jsonKeyword
syn region jsonKeyword matchgroup=jsonQuote start=/"/  end=/"\ze[[:blank:]\r\n]*\:/ contains=tpTemplateExpr contained

" Fix: Same issue with graphqlString — clear and redefine to allow {{ }}
" inside GraphQL string literals.
syn clear graphqlString
syn region graphqlString start=+"+  skip=+\\\\\|\\"+  end=+"\|$+ contains=tpTemplateExpr
syn region graphqlString start=+"""+ skip=+\\"""+ end=+"""+ contains=tpTemplateExpr

" Fix: graphqlFold's start="{" consumes the first { of {{ before
" tpTemplateExpr's start='{{' can match. We clear graphqlFold and redefine
" it with start patterns that skip {{ sequences. We use an explicit contains
" list (instead of ALLBUT) to avoid pulling in tp-specific contained items
" like tpDescLine that would leak into GraphQL fold regions.
syn clear graphqlFold
syn region graphqlFold matchgroup=graphqlBraces start="{\ze[^{]" start="{$" end="}" transparent fold contains=@graphqlSyntax,tpTemplateExpr,tpComment
syn cluster graphqlSyntax add=graphqlFold

syn match  tpTemplateDot      '\.\w\+' contained
syn match  tpTemplatePipe     '|' contained
syn keyword tpTemplateKeyword if else end range with template define block contained
syn match  tpTemplateFunc     '\<\(default\|optional\|timestamp\)\>' contained
syn region tpTemplateString   start='"' end='"' skip='\\"' contained

" --- Body Section with Embedded JSON ---
" The body region spans from 'body:' to EOF or the next top-level key.
syn region tpBody start='^body:' end='^\ze\S' end='\%$' contains=tpBodyKey,@jsonSyntax,tpTemplateExpr
syn match  tpBodyKey   '^body:.*$' contained

" --- Query Section with Embedded GraphQL ---
" The query region spans from 'query:' to EOF or the next top-level key.
" Contains the full GraphQL syntax cluster for proper highlighting.
" Note: matchgroup=tpQueryKey is used on the start pattern so that the
" 'query:' line itself gets Keyword highlighting instead of being matched
" by graphqlStructure (which also recognises the word 'query').
syn region tpQuery matchgroup=tpQueryKey start='^query:.*$' end='^\ze\S' end='\%$' contains=@graphqlSyntax,tpTemplateExpr

" --- Variables Section (YAML-like key: value pairs) ---
" Similar to headers — indented key: value pairs under 'variables:'.
syn region tpVariablesSection start='^variables:' end='^\ze\S' end='\%$' contains=tpVariablesKey,tpVariableLine,tpTemplateExpr
syn match  tpVariablesKey  '^variables:.*$' contained
syn match  tpVariableLine  '^\s\+[A-Za-z0-9_-]\+:.*$' contained contains=tpVarKey,tpVarColon,tpVarValue,tpTemplateExpr
syn match  tpVarKey        '^\s\+\zs[A-Za-z0-9_-]\+\ze:' contained
syn match  tpVarColon      '\%(^\s\+[A-Za-z0-9_-]\+\)\@<=:' contained
syn match  tpVarValue      '\%(^\s\+[A-Za-z0-9_-]\+:\s*\)\@<=.*' contained contains=tpTemplateExpr,tpString

" --- Strings (quoted values) ---
syn region tpString start='"' end='"' skip='\\"' contained contains=tpTemplateExpr
syn region tpString start="'" end="'" skip="\\'" contained

" --- Header lines (contained - only matched inside header context) ---
" These are indented key: value pairs. They are marked 'contained' so they
" do NOT match inside the body region.
syn match tpHeaderLine  '^\s\+[A-Za-z0-9_-]\+:.*$' contained contains=tpHeaderKey,tpHeaderColon,tpHeaderValue,tpTemplateExpr
syn match tpHeaderKey   '^\s\+\zs[A-Za-z0-9_-]\+\ze:' contained
syn match tpHeaderColon '\%(^\s\+[A-Za-z0-9_-]\+\)\@<=:' contained
syn match tpHeaderValue '\%(^\s\+[A-Za-z0-9_-]\+:\s*\)\@<=.*' contained contains=tpTemplateExpr,tpString

" --- Description lines (contained) ---
syn match tpDescLine  '^\s\+\w\+:.*$' contained contains=tpDescKey,tpString,tpTemplateExpr
syn match tpDescKey   '^\s\+\zs\w\+\ze:' contained

" --- Headers key-line: starts a region that contains header lines ---
syn region tpHeaderSection start='^headers:' end='^\ze\S' contains=tpHeaderSectionKey,tpHeaderLine,tpTemplateExpr
syn match  tpHeaderSectionKey '^headers:' contained

" --- Descriptions key-line: starts a region for description lines ---
syn region tpDescSection start='^descriptions:' end='^\ze\S' contains=tpDescSectionKey,tpDescLine,tpTemplateExpr
syn match  tpDescSectionKey '^descriptions:' contained

" --- URL Line ---
syn match tpURLLine    '^url:.*$' contains=tpURLKey,tpURLColon,tpURLValue,tpTemplateExpr
syn match tpURLKey     '^url' contained
syn match tpURLColon   '\%(^url\)\@<=:' contained
syn match tpURLValue   '\%(^url:\s*\)\@<=.\+' contained contains=tpTemplateExpr

" --- Method Line ---
syn match tpMethodLine '^method:.*$' contains=tpMethodKey,tpMethodColon,tpMethod,tpTemplateExpr
syn match tpMethodKey  '^method' contained
syn match tpMethodColon '\%(^method\)\@<=:' contained
syn match tpMethod     '\c\<\(GET\|POST\|PUT\|DELETE\|PATCH\|HEAD\|OPTIONS\|CONNECT\|TRACE\)\>' contained

" --- YAML Comments ---
syn match tpComment '#.*$' containedin=ALL

" --- Highlight Links ---
hi def link tpBodyKey          Keyword
hi def link tpQueryKey         Keyword
hi def link tpVariablesKey     Keyword
hi def link tpHeaderSectionKey Keyword
hi def link tpDescSectionKey   Keyword
hi def link tpURLKey           Keyword
hi def link tpMethodKey        Keyword
hi def link tpURLColon         Delimiter
hi def link tpMethodColon      Delimiter
hi def link tpHeaderColon      Delimiter
hi def link tpVarColon         Delimiter
hi def link tpMethod           Type
hi def link tpURLValue         Underlined
hi def link tpHeaderKey        Identifier
hi def link tpHeaderValue      String
hi def link tpVarKey           Identifier
hi def link tpVarValue         String
hi def link tpString           String
hi def link tpComment          Comment
hi def link tpDescKey          Identifier

" Go template highlighting
hi def link tpTemplateBraces   PreProc
hi def link tpTemplateExpr     PreProc
hi def link tpTemplateDot      Special
hi def link tpTemplateKeyword  Statement
hi def link tpTemplateFunc     Function
hi def link tpTemplateString   String
hi def link tpTemplatePipe     Operator

" --- Sync: force Vim to parse from the start of the file ---
" Without this, multiline regions like tpBody may not be recognised
" when scrolling or opening in the middle of a file.
syn sync fromstart

let b:current_syntax = 'tp'
