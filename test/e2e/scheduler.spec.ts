import { test, expect } from '@playwright/test';

const API_BASE = 'http://localhost:8080/api/v1';
const ADMIN_BASE = 'http://localhost:8080/api/v1/admin';

test.describe('Scheduler E2E Tests', () => {
  let createdJobId: string;

  test.beforeAll(async ({ request }) => {
    // Verify the API is running
    const health = await request.get('http://localhost:8080/health');
    expect(health.ok()).toBeTruthy();
    const healthData = await health.json();
    expect(healthData.status).toBe('up');
  });

  test.afterEach(async ({ request }) => {
    // Clean up created jobs
    if (createdJobId) {
      await request.delete(`${ADMIN_BASE}/jobs/${createdJobId}`);
      createdJobId = null;
    }
  });

  test('should list existing jobs', async ({ request }) => {
    const response = await request.get(`${ADMIN_BASE}/jobs`);
    expect(response.ok()).toBeTruthy();
    
    const jobs = await response.json();
    expect(Array.isArray(jobs)).toBeTruthy();
  });

  test('should create a scheduled job with cron expression', async ({ request }) => {
    const futureDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const jobPayload = {
      name: 'Test Flight Search Job',
      origin: 'JFK',
      destination: 'LAX',
      date_start: futureDate,
      date_end: futureDate,
      adults: 1,
      children: 0,
      infants_lap: 0,
      infants_seat: 0,
      trip_type: 'one_way',
      class: 'economy',
      stops: 'any',
      currency: 'USD',
      interval: 'daily',
      time: '12:00',
      cron_expression: '*/5 * * * *' // Every 5 minutes
    };

    const response = await request.post(`${ADMIN_BASE}/jobs`, {
      data: jobPayload
    });

    expect(response.ok()).toBeTruthy();
    const createdJob = await response.json();
    
    expect(createdJob.id).toBeDefined();
    expect(createdJob.name).toBe(jobPayload.name);
    expect(createdJob.cron_expression).toBe(jobPayload.cron_expression);
    expect(createdJob.trip_type).toBe(jobPayload.trip_type);
    
    createdJobId = createdJob.id;
  });

  test('should retrieve a specific job by ID', async ({ request }) => {
    // First create a job
    const futureDate = new Date(Date.now() + 60 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Get By ID Test Job',
        origin: 'SFO',
        destination: 'ORD',
        date_start: futureDate,
        date_end: futureDate,
        adults: 2,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD',
        interval: 'weekly',
        time: '10:00',
        cron_expression: '0 */6 * * *' // Every 6 hours
      }
    });

    const created = await createResponse.json();
    createdJobId = created.id;

    // Now get it by ID
    const getResponse = await request.get(`${ADMIN_BASE}/jobs/${createdJobId}`);
    expect(getResponse.ok()).toBeTruthy();
    
    const retrievedJob = await getResponse.json();
    expect(retrievedJob.id).toBe(createdJobId);
    expect(retrievedJob.name).toBe('Get By ID Test Job');
    expect(retrievedJob.cron_expression).toBe('0 */6 * * *');
  });

  test('should update an existing job', async ({ request }) => {
    // Create a job first
    const futureDate = new Date(Date.now() + 45 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Update Test Job',
        origin: 'BOS',
        destination: 'MIA',
        date_start: futureDate,
        date_end: futureDate,
        adults: 1,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD',
        interval: 'daily',
        time: '12:00',
        cron_expression: '0 12 * * *' // Daily at noon
      }
    });

    const created = await createResponse.json();
    createdJobId = created.id;

    // Update the job
    const updatePayload = {
      name: 'Updated Test Job',
      origin: 'BOS',
      destination: 'MIA',
      date_start: futureDate,
      date_end: futureDate,
      adults: 2,
      children: 0,
      infants_lap: 0,
      infants_seat: 0,
      trip_type: 'one_way',
      class: 'business',
      stops: 'any',
      currency: 'USD',
      interval: 'daily',
      time: '12:00',
      cron_expression: '0 */12 * * *' // Every 12 hours
    };

    const updateResponse = await request.put(`${ADMIN_BASE}/jobs/${createdJobId}`, {
      data: updatePayload
    });

    expect(updateResponse.ok()).toBeTruthy();
    
    // Verify the update
    const getResponse = await request.get(`${ADMIN_BASE}/jobs/${createdJobId}`);
    const updatedJob = await getResponse.json();
    
    expect(updatedJob.name).toBe('Updated Test Job');
    expect(updatedJob.cron_expression).toBe('0 */12 * * *');
    expect(updatedJob.adults).toBe(2);
    expect(updatedJob.class).toBe('business');
  });

  test('should enable and disable a job', async ({ request }) => {
    // Create a disabled job
    const futureDate = new Date(Date.now() + 21 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Enable/Disable Test Job',
        origin: 'DEN',
        destination: 'SEA',
        date_start: futureDate,
        date_end: futureDate,
        adults: 1,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD',
        interval: 'daily',
        time: '15:00',
        cron_expression: '30 * * * *' // Every hour at 30 minutes
      }
    });

    const created = await createResponse.json();
    createdJobId = created.id;

    // Enable the job
    const enableResponse = await request.post(`${ADMIN_BASE}/jobs/${createdJobId}/enable`);
    expect(enableResponse.ok()).toBeTruthy();

    // Verify it's enabled - check the response status
    expect(enableResponse.status()).toBe(200);

    // Disable the job
    const disableResponse = await request.post(`${ADMIN_BASE}/jobs/${createdJobId}/disable`);
    expect(disableResponse.ok()).toBeTruthy();

    // Verify it's disabled - check the response status
    expect(disableResponse.status()).toBe(200);
  });

  test('should manually run a job', async ({ request }) => {
    // Create a job
    const futureDate = new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Manual Run Test Job',
        origin: 'ATL',
        destination: 'PHX',
        date_start: futureDate,
        date_end: futureDate,
        adults: 1,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD',
        interval: 'weekly',
        time: '09:00',
        cron_expression: '0 0 * * 0' // Weekly on Sunday
      }
    });

    const created = await createResponse.json();
    createdJobId = created.id;

    // Run the job manually
    const runResponse = await request.post(`${ADMIN_BASE}/jobs/${createdJobId}/run`);
    expect(runResponse.ok()).toBeTruthy();
    
    const runResult = await runResponse.json();
    expect(runResult.message).toContain('successfully');
    expect(runResult.job_id).toBeDefined();

    // Wait a bit for the job to be processed
    await new Promise(resolve => setTimeout(resolve, 2000));

    // Check queue status to verify job was added
    const queueResponse = await request.get(`${ADMIN_BASE}/queue`);
    expect(queueResponse.ok()).toBeTruthy();
    
    const queueStatus = await queueResponse.json();
    expect(queueStatus).toHaveProperty('pending_jobs');
    expect(queueStatus).toHaveProperty('processing_jobs');
  });

  test('should handle invalid cron expressions', async ({ request }) => {
    const futureDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const response = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Invalid Cron Test',
        origin: 'LAX',
        destination: 'JFK',
        date_start: futureDate,
        date_end: futureDate,
        adults: 1,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD',
        interval: 'daily',
        time: '14:00',
        cron_expression: 'invalid cron expression'
      }
    });

    expect(response.ok()).toBeFalsy();
    expect(response.status()).toBe(400);
    
    const error = await response.json();
    expect(error.error).toContain('cron');
  });

  test('should delete a job', async ({ request }) => {
    // Create a job
    const futureDate = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Delete Test Job',
        origin: 'ORD',
        destination: 'DFW',
        date_start: futureDate,
        date_end: futureDate,
        adults: 1,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD',
        interval: 'daily',
        time: '16:00',
        cron_expression: '15 */2 * * *' // Every 2 hours at 15 minutes
      }
    });

    const created = await createResponse.json();
    const jobId = created.id;

    // Delete the job
    const deleteResponse = await request.delete(`${ADMIN_BASE}/jobs/${jobId}`);
    expect(deleteResponse.ok()).toBeTruthy();

    // Verify it's deleted
    const getResponse = await request.get(`${ADMIN_BASE}/jobs/${jobId}`);
    expect(getResponse.ok()).toBeFalsy();
    expect(getResponse.status()).toBe(404);

    // No need to clean up since it's already deleted
    createdJobId = null;
  });

  test('should verify scheduler is running via worker status', async ({ request }) => {
    const response = await request.get(`${ADMIN_BASE}/workers`);
    expect(response.ok()).toBeTruthy();
    
    const workerStatus = await response.json();
    expect(workerStatus.workers).toBeGreaterThan(0);
    expect(workerStatus.state).toBe('running');
  });

  test('should handle concurrent job operations', async ({ request }) => {
    // Create multiple jobs concurrently
    const jobPromises = [];
    const jobIds = [];

    for (let i = 0; i < 5; i++) {
      const futureDate = new Date(Date.now() + (30 + i) * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
      const promise = request.post(`${ADMIN_BASE}/jobs`, {
        data: {
          name: `Concurrent Test Job ${i}`,
          origin: 'NYC',
          destination: 'LON',
          date_start: futureDate,
          date_end: futureDate,
          adults: 1,
          children: 0,
          infants_lap: 0,
          infants_seat: 0,
          trip_type: 'one_way',
          class: 'economy',
          stops: 'any',
          currency: 'USD',
          interval: 'daily',
          time: '18:00',
          cron_expression: `${i} * * * *` // Different minute for each
        }
      });
      jobPromises.push(promise);
    }

    const responses = await Promise.all(jobPromises);
    
    for (const response of responses) {
      expect(response.ok()).toBeTruthy();
      const job = await response.json();
      jobIds.push(job.id);
    }

    // Clean up all created jobs
    for (const id of jobIds) {
      await request.delete(`${ADMIN_BASE}/jobs/${id}`);
    }
  });

  test('should validate job execution timing', async ({ request }) => {
    // This test creates a job that should run immediately
    const now = new Date();
    const nextMinute = new Date(now.getTime() + 60000);
    const cronExpression = `${nextMinute.getMinutes()} ${nextMinute.getHours()} * * *`;
    const futureDate = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Timing Test Job',
        origin: 'BOS',
        destination: 'LAX',
        date_start: futureDate,
        date_end: futureDate,
        adults: 1,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'economy',
        stops: 'any',
        currency: 'USD',
        interval: 'daily',
        time: '20:00',
        cron_expression: cronExpression
      }
    });

    const created = await createResponse.json();
    createdJobId = created.id;

    // Get job details to verify next run time
    const getResponse = await request.get(`${ADMIN_BASE}/jobs/${createdJobId}`);
    const job = await getResponse.json();
    
    if (job.next_run_time) {
      const nextRunTime = new Date(job.next_run_time);
      // Verify next run time is within expected range
      expect(nextRunTime.getTime()).toBeGreaterThanOrEqual(nextMinute.getTime() - 1000);
      expect(nextRunTime.getTime()).toBeLessThanOrEqual(nextMinute.getTime() + 1000);
    }
  });
});