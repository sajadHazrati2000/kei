import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ApiService {
  protected http = inject(HttpClient);
  protected base = environment.apiUrl + '/api/v1';
}
