export interface FolderMetadata {
  id: string;
  name: string;
  parentId?: string;
  ownerId: string;
  createdAt: string;
  updatedAt?: string;
}

export interface FileMetadata extends FolderMetadata {
  size: string;
  contentType: string;
}

export interface FolderContents {
  files?: FileMetadata[];
  folders?: FolderMetadata[];
}