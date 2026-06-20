import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TranslateModule } from '@ngx-translate/core';
import { AppShellComponent } from './app-shell.component';

describe('AppShellComponent', () => {
  let fixture: ComponentFixture<AppShellComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        AppShellComponent,
        TranslateModule.forRoot(),
      ],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(AppShellComponent);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('renders the .kei-shell container', () => {
    const shell = fixture.nativeElement.querySelector('.kei-shell');
    expect(shell).withContext('.kei-shell must be present').toBeTruthy();
  });

  it('renders the sidebar', () => {
    const sidebar = fixture.nativeElement.querySelector('app-sidebar');
    expect(sidebar).withContext('app-sidebar must be present').toBeTruthy();
  });

  it('renders the topbar', () => {
    const topbar = fixture.nativeElement.querySelector('app-topbar');
    expect(topbar).withContext('app-topbar must be present').toBeTruthy();
  });

  it('renders the main content area', () => {
    const main = fixture.nativeElement.querySelector('app-main-content');
    expect(main).withContext('app-main-content must be present').toBeTruthy();
  });
});
