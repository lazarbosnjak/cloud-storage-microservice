import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import {
  FolderContents,
  FolderMetadata,
  FileMetadata,
  StorageMetadata,
} from '../models/file-metadata';

@Injectable({
  providedIn: 'root',
})
export class FileService {
  private readonly http = inject(HttpClient);

  private getHeaders() {
    return {
      Authorization: `Bearer ${sessionStorage.getItem('access_token')}`,
    };
  }

  getFolderContents(folderId: string): Observable<FolderContents> {
    return this.http.get<FolderContents>(
      `/api/folders/${encodeURIComponent(folderId)}/contents`,
      {
        headers: this.getHeaders(),
      },
    );
  }

  createFolder(
    name: string,
    parentId: string,
  ): Observable<FolderMetadata> {
    return this.http.post<FolderMetadata>(
      '/api/folders',
      {
        name,
        parentId,
      },
      {
        headers: this.getHeaders(),
      },
    );
  }

  renameFolder(id: string, name: string): Observable<void> {
    return this.http.patch<void>(
      `/api/folders/${encodeURIComponent(id)}/name`,
      {
        name,
      },
      {
        headers: this.getHeaders(),
      },
    );
  }

  renameFile(id: string, name: string): Observable<void> {
    return this.http.patch<void>(
      `/api/files/${encodeURIComponent(id)}/name`,
      {
        name,
      },
      {
        headers: this.getHeaders(),
      },
    );
  }

  uploadFile(file: File, parentId: string): Observable<FileMetadata> {
    return this.http.post<FileMetadata>(
      '/api/files',
      {
        name: file.name,
        size: file.size.toString(),
        contentType: file.type || 'application/octet-stream',
        parentId,
      },
      {
        headers: this.getHeaders(),
      },
    );
  }

  getStorage(): Observable<StorageMetadata> {
  return this.http.get<StorageMetadata>(
    '/api/users/storage',
      {
       headers: this.getHeaders(),
      },
    );
  }
}