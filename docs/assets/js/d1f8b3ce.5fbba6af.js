"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[353],{5792:(e,t,s)=>{s.r(t),s.d(t,{assets:()=>a,contentTitle:()=>o,default:()=>l,frontMatter:()=>u,metadata:()=>r,toc:()=>d});const r=JSON.parse('{"id":"Database/querybuilder","title":"Query Builder","description":"While dbutils CRUD helper functions are useful for simple database operations, they may not be sufficient for complex queries or when you need to build dynamic queries. For these cases, you can use the querybuilder which supports joins, complex where clauses, group by, order by, limit, and offset.","source":"@site/docs/Database/querybuilder.md","sourceDirName":"Database","slug":"/Database/querybuilder","permalink":"/gowebutils/docs/Database/querybuilder","draft":false,"unlisted":false,"editUrl":"https://github.com/gurch101/gowebutils/tree/main/packages/create-docusaurus/templates/shared/docs/Database/querybuilder.md","tags":[],"version":"current","sidebarPosition":3,"frontMatter":{"sidebar_position":3},"sidebar":"tutorialSidebar","previous":{"title":"CRUD Helpers","permalink":"/gowebutils/docs/Database/utilities"},"next":{"title":"Transactions","permalink":"/gowebutils/docs/Database/transactions"}}');var n=s(4848),i=s(8453);const u={sidebar_position:3},o="Query Builder",a={},d=[{value:"Usage",id:"usage",level:2}];function c(e){const t={code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",...(0,i.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(t.header,{children:(0,n.jsx)(t.h1,{id:"query-builder",children:"Query Builder"})}),"\n",(0,n.jsxs)(t.p,{children:["While ",(0,n.jsx)(t.code,{children:"dbutils"})," CRUD helper functions are useful for simple database operations, they may not be sufficient for complex queries or when you need to build dynamic queries. For these cases, you can use the ",(0,n.jsx)(t.code,{children:"querybuilder"})," which supports joins, complex where clauses, group by, order by, limit, and offset."]}),"\n",(0,n.jsx)(t.h2,{id:"usage",children:"Usage"}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-go",children:'// runs SELECT id, name, age FROM users WHERE (age > 18) OR (name LIKE "%doe%") LIMIT 10 OFFSET 10\nquerybuilder := dbutils.NewQueryBuilder(db).\n\tSelect("id", "name", "age").\n\tFrom("users").\n\tWhere("age > ?", 18).\n\tOrWhereLike("name", dbutils.OpContains, "doe").\n\tLimit(10).\n\tOffset(10).\n\tQuery(func(rows *sql.Rows) error {\n\t\t// do something with rows\n\t})\n'})}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-go",children:'// runs SELECT u.id, COUNT(c.comments) FROM users u INNER JOIN comments c ON u.id = c.user_id WHERE (u.age > 18) AND (u.active) OR (u.name = "doe") GROUP BY u.id ORDER BY u.name DESC LIMIT 10 OFFSET 20\nqueryBuilder := dbutils.NewQueryBuilder(db).\n\t\tSelect("u.id", "COUNT(c.comments)").\n\t\tFrom("users u").\n\t\tJoin("INNER", "comments c", "u.id = c.user_id").\n\t\tWhere("u.age > ?", 18).\n\t\tAndWhere("u.active = ?", true).\n\t\tOrWhere("u.name = ?", "doe").\n\t\tGroupBy("u.id").\n\t\t// -<fieldname> is used for descending order\n\t\t// <fieldname> is used for ascending order\n\t\tOrderBy("-u.name").\n\t\tLimit(10).\n\t\tOffset(20).\n\t\tQuery(func(rows *sql.Rows) error {\n\t\t\t// do something with rows\n\t\t})\n'})}),"\n",(0,n.jsx)(t.p,{children:"nils values passed to any of the where clause functions will be ignored which means that you can avoid branching in your code."}),"\n",(0,n.jsxs)(t.p,{children:["Rather than using ",(0,n.jsx)(t.code,{children:"Exec"}),", you can also use the QueryBuilder to get the query string and arguments by calling the ",(0,n.jsx)(t.code,{children:"Build()"})," method."]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-go",children:'\tqb := dbutils.NewQueryBuilder(db).Select("id", "name").From("users").Where("id = ?", 1)\n\t// returns SELECT id, name FROM users WHERE (id = ?) and []{1}\n\tquery, args := qb.Build()\n'})})]})}function l(e={}){const{wrapper:t}={...(0,i.R)(),...e.components};return t?(0,n.jsx)(t,{...e,children:(0,n.jsx)(c,{...e})}):c(e)}},8453:(e,t,s)=>{s.d(t,{R:()=>u,x:()=>o});var r=s(6540);const n={},i=r.createContext(n);function u(e){const t=r.useContext(i);return r.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function o(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(n):e.components||n:u(e.components),r.createElement(i.Provider,{value:t},e.children)}}}]);