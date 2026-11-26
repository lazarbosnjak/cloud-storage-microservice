import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { FolderContents } from '../models/file-metadata';

@Injectable({
  providedIn: 'root',
})
export class FileService {
  private readonly http = inject(HttpClient);

  getFolderContents(folderId: string): Observable<FolderContents> {
    const token = sessionStorage.getItem('access_token');

    return this.http.get<FolderContents>(
      `/api/folders/${encodeURIComponent(folderId)}/contents`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      },
    );
  }
}