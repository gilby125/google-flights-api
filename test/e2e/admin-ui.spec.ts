import { test, expect } from '@playwright/test';

test.describe('Admin UI Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/admin/');
    await expect(page).toHaveTitle('Admin Panel - Google Flights API');
  });

  test('should load admin dashboard', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Dashboard');
    await expect(page.locator('#jobsCount')).toBeVisible();
    await expect(page.locator('#workersCount')).toBeVisible();
    await expect(page.locator('#queueSize')).toBeVisible();
  });

  test('should create a new job through UI', async ({ page }) => {
    // Click New Job button
    await page.click('button:has-text("New Job")');
    
    // Fill out the form
    await page.fill('#jobName', 'UI Test Job');
    await page.fill('#origin', 'JFK');
    await page.fill('#destination', 'LAX');
    
    // The dates should be auto-filled, but let's verify
    const dateStart = await page.inputValue('#dateStart');
    const dateEnd = await page.inputValue('#dateEnd');
    expect(dateStart).toBeTruthy();
    expect(dateEnd).toBeTruthy();
    
    // Select trip type and class
    await page.selectOption('#tripType', 'one_way');
    await page.selectOption('#class', 'business');
    await page.selectOption('#stops', 'nonstop');
    
    // Set passenger counts
    await page.fill('#adults', '2');
    await page.fill('#children', '1');
    
    // Select currency
    await page.selectOption('#currency', 'EUR');
    
    // Change schedule settings
    await page.selectOption('#interval', 'weekly');
    await page.fill('#time', '15:30');
    
    // Verify cron preview updates
    await expect(page.locator('#cronPreview')).toContainText('30 15 * * 1');
    
    // Save the job
    await page.click('#saveJobBtn');
    
    // Wait for success message
    await expect(page.locator('.alert-success')).toContainText('Job created successfully!');
    
    // Verify job appears in the table
    await expect(page.locator('#jobsTable')).toContainText('UI Test Job');
    await expect(page.locator('#jobsTable')).toContainText('JFK → LAX');
  });

  test('should show cron preview correctly', async ({ page }) => {
    await page.click('button:has-text("New Job")');
    
    // Test daily schedule
    await page.selectOption('#interval', 'daily');
    await page.fill('#time', '09:15');
    await expect(page.locator('#cronPreview')).toContainText('15 9 * * *');
    
    // Test weekly schedule
    await page.selectOption('#interval', 'weekly');
    await page.fill('#time', '14:45');
    await expect(page.locator('#cronPreview')).toContainText('45 14 * * 1');
    
    // Test monthly schedule
    await page.selectOption('#interval', 'monthly');
    await page.fill('#time', '08:00');
    await expect(page.locator('#cronPreview')).toContainText('0 8 1 * *');
  });

  test('should handle form validation', async ({ page }) => {
    await page.click('button:has-text("New Job")');
    
    // Try to save empty form
    await page.click('#saveJobBtn');
    
    // Should show browser validation errors (form should not submit)
    const modalTitle = await page.locator('.modal-title');
    await expect(modalTitle).toContainText('Create New Scheduled Job');
  });

  test('should display existing jobs', async ({ page }) => {
    // Wait for jobs to load
    await page.waitForSelector('#jobsTable tr', { timeout: 5000 });
    
    // Check that jobs are displayed
    const jobRows = await page.locator('#jobsTable tr').count();
    expect(jobRows).toBeGreaterThan(0);
    
    // Check for action buttons
    await expect(page.locator('#jobsTable button[onclick*="runJob"]')).toBeVisible();
    await expect(page.locator('#jobsTable button[onclick*="deleteJob"]')).toBeVisible();
  });

  test('should run a job manually', async ({ page }) => {
    // Wait for jobs to load
    await page.waitForSelector('#jobsTable tr', { timeout: 5000 });
    
    // Click the first run button
    const runButton = page.locator('#jobsTable button[onclick*="runJob"]').first();
    await runButton.click();
    
    // Wait for success message
    await expect(page.locator('.alert-success')).toContainText('Job started successfully!');
  });
});