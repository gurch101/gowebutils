"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[361],{7585:(n,e,t)=>{t.r(e),t.d(e,{assets:()=>c,contentTitle:()=>o,default:()=>d,frontMatter:()=>i,metadata:()=>a,toc:()=>l});const a=JSON.parse('{"id":"Database/transactions","title":"Transactions","description":"To use transactions, use app.DB.WithTransaction to wrap your database operations in a transaction. Transactions can be nested simply by calling WithTransaction again. If an error is returned by your callback function, the transaction will be rolled back.","source":"@site/docs/Database/transactions.md","sourceDirName":"Database","slug":"/Database/transactions","permalink":"/gowebutils/docs/Database/transactions","draft":false,"unlisted":false,"editUrl":"https://github.com/gurch101/gowebutils/tree/main/packages/create-docusaurus/templates/shared/docs/Database/transactions.md","tags":[],"version":"current","sidebarPosition":4,"frontMatter":{"sidebar_position":4},"sidebar":"tutorialSidebar","previous":{"title":"Query Builder","permalink":"/gowebutils/docs/Database/querybuilder"},"next":{"title":"Testing","permalink":"/gowebutils/docs/Database/testing"}}');var r=t(4848),s=t(8453);const i={sidebar_position:4},o="Transactions",c={},l=[];function u(n){const e={code:"code",h1:"h1",header:"header",p:"p",pre:"pre",...(0,s.R)(),...n.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(e.header,{children:(0,r.jsx)(e.h1,{id:"transactions",children:"Transactions"})}),"\n",(0,r.jsxs)(e.p,{children:["To use transactions, use ",(0,r.jsx)(e.code,{children:"app.DB.WithTransaction"})," to wrap your database operations in a transaction. Transactions can be nested simply by calling ",(0,r.jsx)(e.code,{children:"WithTransaction"})," again. If an error is returned by your callback function, the transaction will be rolled back."]}),"\n",(0,r.jsx)(e.pre,{children:(0,r.jsx)(e.code,{className:"language-go",children:'err := dbutils.WithTransaction(ctx, db, func(tx dbutils.DB) error {\n  tenantID, err := dbutils.Insert(ctx, tx, "tenants", map[string]any{\n    "tenant_name":   uuid.New().String(),\n    "contact_email": email,\n    "plan":          "free",\n  })\n\n  if err != nil {\n    return fmt.Errorf("failed to create tenant: %w", err)\n  }\n\n  userID, err = dbutils.Insert(ctx, tx, "users", map[string]any{\n    "tenant_id": tenantID,\n    "user_name": username,\n    "email":     email,\n  })\n\n  if err != nil {\n    return fmt.Errorf("failed to create user: %w", err)\n  }\n\n  return nil\n})\n'})})]})}function d(n={}){const{wrapper:e}={...(0,s.R)(),...n.components};return e?(0,r.jsx)(e,{...n,children:(0,r.jsx)(u,{...n})}):u(n)}},8453:(n,e,t)=>{t.d(e,{R:()=>i,x:()=>o});var a=t(6540);const r={},s=a.createContext(r);function i(n){const e=a.useContext(s);return a.useMemo((function(){return"function"==typeof n?n(e):{...e,...n}}),[e,n])}function o(n){let e;return e=n.disableParentContext?"function"==typeof n.components?n.components(r):n.components||r:i(n.components),a.createElement(s.Provider,{value:e},n.children)}}}]);