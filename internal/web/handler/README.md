# 文件上传接口调用文档

## 概述

本文档详细说明了文件上传相关接口的调用方法和顺序，特别是大文件分片上传及断点续传功能的实现流程。

## 核心接口

系统提供了以下核心文件上传相关接口：

1. 初始化文件上传
2. 上传文件分片
3. 获取上传状态
4. 断点续传
5. 完成文件上传

## 10GB文件上传流程

对于10GB的大文件`ceshi.txt`，完整的上传流程如下：

### 1. 初始化文件上传

**接口信息**
- URL: `/api/v1/files/init`
- 方法: POST
- 说明: 初始化文件上传，获取文件ID和分片信息

**请求参数**
```json
{
  "fileName": "ceshi.txt",          // 文件名
  "fileSize": 10737418240,          // 文件大小（10GB）
  "chunkSize": 10485760,            // 分片大小（建议10MB，最大10MB）
  "fileHash": "可选的文件哈希值",   // 可选，用于完整性校验
  "fileType": "text/plain"          // 文件类型
}
```

**响应示例**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 123,                   // 文件ID，后续操作需要使用
    "upload_id": "uuid-123456",      // 上传ID，断点续传时需要使用
    "total_chunks": 1025,             // 总分片数（自动计算）
    "chunk_size": 10485760,           // 分片大小
    "upload_path": "uploads/temp/uuid-123456" // 临时上传路径
  }
}
```

### 2. 上传文件分片

**接口信息**
- URL: `/api/v1/files/chunk`
- 方法: POST
- 说明: 上传文件分片，支持大文件的分片上传

**请求参数（Form表单）**
- fileId: 文件ID（从初始化接口获取）
- chunkIndex: 分片索引（从0开始）
- totalChunks: 总分片数
- chunkHash: 可选，分片哈希值
- file: 分片文件数据

**响应示例**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 123,
    "chunk_index": 0,
    "uploaded_chunks": 1,
    "total_chunks": 1025,
    "progress": 0.0976
  }
}
```

### 3. 获取上传状态（可选）

在上传过程中，可以随时查询文件的上传状态：

**接口信息**
- URL: `/api/v1/files/status?fileId=123`
- 方法: GET
- 说明: 查询文件上传状态，获取已上传分片信息

**响应示例**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 123,
    "file_name": "ceshi.txt",
    "file_size": 10737418240,
    "total_chunks": 1025,
    "uploaded_chunks": 500,
    "chunk_indices": [0, 1, 2, ..., 499],
    "progress": 48.78,
    "status": "uploading",
    "upload_id": "uuid-123456",
    "create_time": "2023-07-01T10:00:00Z"
  }
}
```

### 4. 断点续传

如果上传中断，客户端可以使用断点续传功能继续上传：

**接口信息**
- URL: `/api/v1/files/resume?uploadId=uuid-123456`
- 方法: GET
- 说明: 根据uploadId获取已上传的分片信息，继续上传

**响应示例**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 123,
    "file_name": "ceshi.txt",
    "file_size": 10737418240,
    "total_chunks": 1025,
    "chunk_size": 10485760,
    "uploaded_chunks": 500,
    "chunk_indices": [0, 1, 2, ..., 499],
    "progress": 48.78,
    "status": "uploading",
    "upload_id": "uuid-123456",
    "create_time": "2023-07-01T10:00:00Z"
  }
}
```

**断点续传流程**:
1. 调用resume接口，获取已上传的分片索引列表
2. 根据返回的已上传分片索引，跳过这些分片，只上传未上传的分片
3. 继续使用uploadChunk接口上传剩余的分片

### 5. 完成文件上传

当所有分片上传完成后，调用完成上传接口：

**接口信息**
- URL: `/api/v1/files/complete`
- 方法: POST
- 说明: 通知服务器文件分片上传完成，进行文件合并

**请求参数**
```json
{
  "fileId": 123,                    // 文件ID
  "fileHash": "可选的文件哈希值",  // 可选，用于完整性校验
  "uploader": "可选的上传者信息"   // 可选，上传者信息
}
```

**响应示例**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 123,
    "file_name": "ceshi.txt",
    "file_size": 10737418240,
    "file_hash": "file-hash-value",
    "file_path": "uploads/ceshi_20230701_101530.txt",
    "download_url": "/api/v1/files/123/download",
    "upload_time": "2023-07-01T10:00:00Z",
    "complete_time": "2023-07-01T10:15:30Z"
  }
}
```

## 完整调用示例

### 客户端实现逻辑

```javascript
// 1. 初始化上传
async function initUpload(file) {
  const chunkSize = 10 * 1024 * 1024; // 10MB分片
  const response = await fetch('/api/v1/files/init', {
    method: 'POST',
    body: JSON.stringify({
      fileName: file.name,
      fileSize: file.size,
      chunkSize: chunkSize,
      fileType: file.type
    })
  });
  const data = await response.json();
  return data.data;
}

// 2. 上传分片
async function uploadChunk(file, fileId, chunkIndex, totalChunks) {
  const chunkSize = 10 * 1024 * 1024;
  const start = chunkIndex * chunkSize;
  const end = Math.min(start + chunkSize, file.size);
  const chunk = file.slice(start, end);
  
  const formData = new FormData();
  formData.append('fileId', fileId);
  formData.append('chunkIndex', chunkIndex);
  formData.append('totalChunks', totalChunks);
  formData.append('file', chunk);
  
  const response = await fetch('/api/v1/files/chunk', {
    method: 'POST',
    body: formData
  });
  
  return response.json();
}

// 3. 断点续传逻辑
async function resumeUpload(uploadId) {
  const response = await fetch(`/api/v1/files/resume?uploadId=${uploadId}`);
  const data = await response.json();
  return data.data;
}

// 4. 完成上传
async function completeUpload(fileId) {
  const response = await fetch('/api/v1/files/complete', {
    method: 'POST',
    body: JSON.stringify({ fileId })
  });
  
  return response.json();
}

// 5. 主上传函数
async function uploadLargeFile(file) {
  try {
    // 尝试恢复上传
    // 如果有保存的uploadId，可以先用它来尝试恢复上传
    let uploadInfo = null;
    let uploadId = localStorage.getItem('uploadId');
    
    if (uploadId) {
      try {
        uploadInfo = await resumeUpload(uploadId);
        console.log('恢复上传成功:', uploadInfo);
      } catch (error) {
        console.log('恢复上传失败，开始新的上传');
        // 初始化上传
        uploadInfo = await initUpload(file);
        localStorage.setItem('uploadId', uploadInfo.upload_id);
      }
    } else {
      // 初始化上传
      uploadInfo = await initUpload(file);
      localStorage.setItem('uploadId', uploadInfo.upload_id);
    }
    
    const { file_id, total_chunks, uploaded_chunk_indexes } = uploadInfo;
    
    // 上传所有未上传的分片
    for (let i = 0; i < total_chunks; i++) {
      // 跳过已上传的分片
      if (uploaded_chunk_indexes && uploaded_chunk_indexes.includes(i)) {
        continue;
      }
      
      const result = await uploadChunk(file, file_id, i, total_chunks);
      console.log(`上传分片 ${i + 1}/${total_chunks} 完成，进度: ${result.data.progress.toFixed(2)}%`);
      
      // 可以在这里更新UI进度
    }
    
    // 完成上传
    const finalResult = await completeUpload(file_id);
    console.log('文件上传完成:', finalResult);
    
    // 清除保存的uploadId
    localStorage.removeItem('uploadId');
    
    return finalResult;
  } catch (error) {
    console.error('文件上传失败:', error);
    throw error;
  }
}

// 使用示例
const fileInput = document.querySelector('input[type="file"]');
fileInput.addEventListener('change', async (event) => {
  const file = event.target.files[0];
  if (file) {
    await uploadLargeFile(file);
  }
});
```

## 注意事项

1. **分片大小限制**：分片大小最大为10MB，超过会被服务端拒绝

2. **分片索引**：分片索引从0开始，必须连续且不超过总分片数

3. **断点续传**：
   - 客户端需要保存服务端返回的`upload_id`以便在需要时恢复上传
   - 建议将`upload_id`保存在本地存储或会话存储中

4. **网络异常处理**：
   - 上传过程中应处理网络异常，可通过重试机制提高成功率
   - 重试前可以先调用`resume`接口获取已上传的分片信息

5. **文件校验**：
   - 可选地计算文件哈希值，在完成上传时提供给服务端进行校验
   - 分片上传时也可以提供分片哈希值，提高传输可靠性

6. **大文件处理**：
   - 对于10GB这样的大文件，建议实现分片上传进度的可视化显示
   - 考虑实现暂停/恢复功能，提高用户体验

7. **服务器处理**：
   - 服务端会在完成上传时合并所有分片
   - 合并完成后会删除临时分片文件

8. **错误处理**：
   - 遇到错误时，服务端会返回具体的错误码和错误信息
   - 客户端应根据错误信息采取相应措施

通过以上接口和流程，可以高效、可靠地实现大文件的分片上传和断点续传功能。