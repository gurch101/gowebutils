"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[423],{5322:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>c,contentTitle:()=>a,default:()=>u,frontMatter:()=>r,metadata:()=>i,toc:()=>d});const i=JSON.parse('{"id":"Database/intro","title":"Intro","description":"gowebutils is optimized for use with SQLite databases. When initializing your App, a connection pool with 1 connection is created that supports writes and 10 connections are created that support reads. When issuing writes, simply use the App.DB.WithTransaction function to execute your write queries. Connections are created with foreign keys and WAL mode enabled.","source":"@site/docs/Database/intro.md","sourceDirName":"Database","slug":"/Database/intro","permalink":"/gowebutils/docs/Database/intro","draft":false,"unlisted":false,"editUrl":"https://github.com/gurch101/gowebutils/tree/main/packages/create-docusaurus/templates/shared/docs/Database/intro.md","tags":[],"version":"current","sidebarPosition":1,"frontMatter":{"sidebar_position":1},"sidebar":"tutorialSidebar","previous":{"title":"Application","permalink":"/gowebutils/docs/app"},"next":{"title":"CRUD Helpers","permalink":"/gowebutils/docs/Database/utilities"}}');var s=n(4848),o=n(8453);const r={sidebar_position:1},a="Intro",c={},d=[{value:"Intialization",id:"intialization",level:3}];function l(e){const t={code:"code",h1:"h1",h3:"h3",header:"header",p:"p",pre:"pre",...(0,o.R)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(t.header,{children:(0,s.jsx)(t.h1,{id:"intro",children:"Intro"})}),"\n",(0,s.jsxs)(t.p,{children:[(0,s.jsx)(t.code,{children:"gowebutils"})," is optimized for use with SQLite databases. When initializing your ",(0,s.jsx)(t.code,{children:"App"}),", a connection pool with 1 connection is created that supports writes and 10 connections are created that support reads. When issuing writes, simply use the ",(0,s.jsx)(t.code,{children:"App.DB.WithTransaction"})," function to execute your write queries. Connections are created with foreign keys and WAL mode enabled."]}),"\n",(0,s.jsxs)(t.p,{children:["The ",(0,s.jsx)(t.code,{children:"App.DB"})," object shares a similar interface to the ",(0,s.jsx)(t.code,{children:"sql.DB"})," object and has the following methods: - ",(0,s.jsx)(t.code,{children:"ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)"})," - ",(0,s.jsx)(t.code,{children:"QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)"})," - ",(0,s.jsx)(t.code,{children:"QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row"})]}),"\n",(0,s.jsx)(t.h3,{id:"intialization",children:"Intialization"}),"\n",(0,s.jsx)(t.p,{children:"A single environment variable is required to initialize your database."}),"\n",(0,s.jsx)(t.pre,{children:(0,s.jsx)(t.code,{className:"language-sh",children:'# The sqlite3 database file path\nexport DB_FILEPATH="./app.db"\n'})})]})}function u(e={}){const{wrapper:t}={...(0,o.R)(),...e.components};return t?(0,s.jsx)(t,{...e,children:(0,s.jsx)(l,{...e})}):l(e)}},8453:(e,t,n)=>{n.d(t,{R:()=>r,x:()=>a});var i=n(6540);const s={},o=i.createContext(s);function r(e){const t=i.useContext(o);return i.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function a(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:r(e.components),i.createElement(o.Provider,{value:t},e.children)}}}]);