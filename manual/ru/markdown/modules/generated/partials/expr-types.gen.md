[cols="2m,3a"]
|===
| Type | Structure

| [#objref-attachment]#[objref-attachment,Attachment](#objref-attachment,Attachment)#
|
```
```
{
   ID (ID)
   kind (String)
   url (Handle)
   previewUrl (Handle)
   name (Handle)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-composemodule]#[objref-composemodule,ComposeModule](#objref-composemodule,ComposeModule)#
|
```
```
{
   ID (ID)
   namespaceID (ID)
   name (String)
   handle (Handle)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-composenamespace]#[objref-composenamespace,ComposeNamespace](#objref-composenamespace,ComposeNamespace)#
|
```
```
{
   ID (ID)
   name (String)
   slug (Handle)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-composerecord]#[objref-composerecord,ComposeRecord](#objref-composerecord,ComposeRecord)#
|
```
```
{
   ID (ID)
   moduleID (ID)
   namespaceID (ID)
   values (ComposeRecordValues)
   meta (Meta)
   ownedBy (ID)
   createdAt (DateTime)
   createdBy (ID)
   updatedAt (DateTime)
   updatedBy (ID)
   deletedAt (DateTime)
   deletedBy (ID)
}


| [#objref-httprequest]#[objref-httprequest,HttpRequest](#objref-httprequest,HttpRequest)#
|
```
```
{
   Method (String)
   URL (Url)
   Header (KVV)
   Body (Reader)
   Form (KVV)
   PostForm (KVV)
}


| [#objref-queuemessage]#[objref-queuemessage,QueueMessage](#objref-queuemessage,QueueMessage)#
|
```
```
{
   Queue (String)
   Payload (Bytes)
}

| [#objref-rendereddocument]#[objref-rendereddocument,RenderedDocument](#objref-rendereddocument,RenderedDocument)#
|
```
```
{
   document (Reader)
   name (string)
   type (string)
}

| [#objref-role]#[objref-role,Role](#objref-role,Role)#
|
```
```
{
   ID (ID)
   name (String)
   handle (Handle)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   archivedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-template]#[objref-template,Template](#objref-template,Template)#
|
```
```
{
   ID (ID)
   handle (Handle)
   language (String)
   type (DocumentType)
   partial (Boolean)
   meta (TemplateMeta)
   template (String)
   labels (KV)
   ownerID (ID)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
   lastUsedAt (DateTime)
}

| [#objref-templatemeta]#[objref-templatemeta,TemplateMeta](#objref-templatemeta,TemplateMeta)#
|
```
```
{
   short (String)
   description (String)
}

| [#objref-user]#[objref-user,User](#objref-user,User)#
|
```
```
{
   ID (ID)
   username (String)
   email (String)
   name (String)
   handle (Handle)
   emailConfirmed (Boolean)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   suspendedAt (DateTime)
   deletedAt (DateTime)
}


|===
