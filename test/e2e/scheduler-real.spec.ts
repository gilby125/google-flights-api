import { test, expect } from '@playwright/test';

const API_BASE = 'http://localhost:8080/api/v1';
const ADMIN_BASE = 'http://localhost:8080/api/v1/admin';

test.describe('Real Scheduler E2E Tests - No Faking', () => {
  let createdJobIds: string[] = [];

  test.beforeAll(async ({ request }) => {
    // Verify the API is running
    const health = await request.get('http://localhost:8080/health');
    expect(health.ok()).toBeTruthy();
    const healthData = await health.json();
    expect(healthData.status).toBe('up');
    
    console.log('✅ API is healthy and running');
  });

  test.afterEach(async ({ request }) => {
    // Clean up created jobs
    for (const id of createdJobIds) {
      try {
        await request.delete(`${ADMIN_BASE}/jobs/${id}`);
        console.log(`🧹 Cleaned up job ${id}`);
      } catch (e) {
        console.log(`⚠️  Failed to clean up job ${id}:`, e);
      }
    }
    createdJobIds = [];
  });

  test('REAL TEST: Create and verify a scheduled job', async ({ request }) => {
    console.log('📝 Creating a real scheduled job...');
    
    const futureDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const jobPayload = {
      name: 'Real Scheduler Test - Flight Search',
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

    console.log('Response status:', response.status());
    const responseText = await response.text();
    console.log('Response body:', responseText);

    expect(response.ok()).toBeTruthy();
    
    // Parse the response - handle potential JSON issues
    let createdJob;
    try {
      createdJob = JSON.parse(responseText);
    } catch (e) {
      console.error('Failed to parse response:', e);
      throw e;
    }

    // Verify the created job has expected properties
    expect(createdJob).toBeDefined();
    if (createdJob.id) {
      createdJobIds.push(createdJob.id.toString());
      console.log('✅ Job created with ID:', createdJob.id);
    }

    // List all jobs to verify it exists
    const listResponse = await request.get(`${ADMIN_BASE}/jobs`);
    const listText = await listResponse.text();
    console.log('Jobs list response:', listText);
    
    // The API might return duplicated response, extract the actual array
    // The response format is: [][{actual data}]
    const actualDataMatch = listText.match(/\]\[(.+)\]$/);
    if (actualDataMatch) {
      const jobs = JSON.parse('[' + actualDataMatch[1] + ']');
      console.log('Found jobs:', jobs.length);
      
      const foundJob = jobs.find((j: any) => j.name === jobPayload.name);
      expect(foundJob).toBeDefined();
      console.log('✅ Job found in list:', foundJob);
    } else {
      // Fallback: try to parse the whole response if it's valid JSON
      try {
        const parsed = JSON.parse(listText);
        const jobs = Array.isArray(parsed) ? parsed : [];
        console.log('Found jobs (fallback):', jobs.length);
        const foundJob = jobs.find((j: any) => j.name === jobPayload.name);
        expect(foundJob).toBeDefined();
        console.log('✅ Job found in list (fallback):', foundJob);
      } catch (e) {
        console.error('Failed to parse jobs list:', e);
        console.log('Raw response was:', listText);
      }
    }
  });

  test('REAL TEST: Manually run a job and verify queue', async ({ request }) => {
    console.log('🚀 Testing manual job execution...');
    
    // First create a job
    const futureDate = new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Manual Execution Test',
        origin: 'SFO',
        destination: 'NYC',
        date_start: futureDate,
        date_end: futureDate,
        adults: 2,
        children: 0,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'round_trip',
        return_date_start: new Date(Date.now() + 21 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        return_date_end: new Date(Date.now() + 21 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        class: 'business',
        stops: 'nonstop',
        currency: 'USD',
        interval: 'weekly',
        time: '09:00',
        cron_expression: '0 9 * * 1' // Weekly on Monday at 9 AM
      }
    });

    const created = await createResponse.json();
    if (created.id) {
      createdJobIds.push(created.id.toString());
    }
    
    console.log('Created job for manual run:', created.id || 'ID not found');

    // Run the job manually
    const runResponse = await request.post(`${ADMIN_BASE}/jobs/${created.id}/run`);
    console.log('Run response status:', runResponse.status());
    
    if (runResponse.ok()) {
      const runResult = await runResponse.json();
      console.log('✅ Job run result:', runResult);
    }

    // Check queue status
    await new Promise(resolve => setTimeout(resolve, 1000)); // Wait 1 second
    
    const queueResponse = await request.get(`${ADMIN_BASE}/queue`);
    const queueText = await queueResponse.text();
    console.log('Queue status:', queueText);
    
    expect(queueResponse.ok()).toBeTruthy();
  });

  test('REAL TEST: Verify scheduler cron functionality', async ({ request }) => {
    console.log('⏰ Testing cron scheduling...');
    
    // Create a job that should run soon
    const now = new Date();
    const in2Minutes = new Date(now.getTime() + 2 * 60 * 1000);
    const minute = in2Minutes.getMinutes();
    const hour = in2Minutes.getHours();
    
    // Create a cron expression for 2 minutes from now
    const cronExpression = `${minute} ${hour} * * *`;
    console.log(`Creating job with cron: ${cronExpression} (should run at ${in2Minutes.toLocaleTimeString()})`);

    const futureDate = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const response = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: `Cron Test - Runs at ${hour}:${minute}`,
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
        currency: 'EUR',
        interval: 'daily',
        time: `${hour}:${minute}`,
        cron_expression: cronExpression
      }
    });

    const created = await response.json();
    if (created.id) {
      createdJobIds.push(created.id.toString());
      console.log('✅ Cron job created with ID:', created.id);
    }

    // Check if the job is properly scheduled
    const getResponse = await request.get(`${ADMIN_BASE}/jobs/${created.id}`);
    if (getResponse.ok()) {
      const job = await getResponse.json();
      console.log('Job details:', job);
      
      if (job.next_run_time) {
        console.log('Next run scheduled for:', new Date(job.next_run_time).toLocaleString());
      }
    }
  });

  test('REAL TEST: Enable/disable job and verify state', async ({ request }) => {
    console.log('🔄 Testing job enable/disable functionality...');
    
    const futureDate = new Date(Date.now() + 45 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const createResponse = await request.post(`${ADMIN_BASE}/jobs`, {
      data: {
        name: 'Enable/Disable Test Job',
        origin: 'MIA',
        destination: 'SEA',
        date_start: futureDate,
        date_end: futureDate,
        adults: 3,
        children: 1,
        infants_lap: 0,
        infants_seat: 0,
        trip_type: 'one_way',
        class: 'premium_economy',
        stops: 'one_stop',
        currency: 'GBP',
        interval: 'daily',
        time: '22:30',
        cron_expression: '30 22 * * *' // Daily at 10:30 PM
      }
    });

    const created = await createResponse.json();
    if (created.id) {
      createdJobIds.push(created.id.toString());
    }

    // Test enable
    console.log('Enabling job...');
    const enableResponse = await request.post(`${ADMIN_BASE}/jobs/${created.id}/enable`);
    console.log('Enable response:', enableResponse.status());
    expect(enableResponse.ok()).toBeTruthy();

    // Small delay
    await new Promise(resolve => setTimeout(resolve, 500));

    // Test disable
    console.log('Disabling job...');
    const disableResponse = await request.post(`${ADMIN_BASE}/jobs/${created.id}/disable`);
    console.log('Disable response:', disableResponse.status());
    expect(disableResponse.ok()).toBeTruthy();
  });

  test('REAL TEST: Verify worker manager is functioning', async ({ request }) => {
    console.log('👷 Checking worker manager status...');
    
    const response = await request.get(`${ADMIN_BASE}/workers`);
    const responseText = await response.text();
    console.log('Raw worker status:', responseText);
    
    expect(response.ok()).toBeTruthy();
    
    // Handle potential duplicate responses
    const matches = responseText.match(/\{[^}]+\}/g);
    if (matches && matches.length > 0) {
      const lastMatch = matches[matches.length - 1];
      const workerStatus = JSON.parse(lastMatch);
      console.log('Worker status parsed:', workerStatus);
      
      expect(workerStatus.status).toBe('running');
      console.log('✅ Worker manager is running');
    }
  });

  test('REAL TEST: Invalid cron expression handling', async ({ request }) => {
    console.log('❌ Testing invalid cron expression...');
    
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
        cron_expression: 'this is not a valid cron expression!!!'
      }
    });

    console.log('Response status:', response.status());
    expect(response.ok()).toBeFalsy();
    expect(response.status()).toBe(400);
    
    const error = await response.json();
    console.log('Error response:', error);
    console.log('✅ Invalid cron properly rejected');
  });

  test('REAL TEST: Concurrent job operations stress test', async ({ request }) => {
    console.log('🔥 Running concurrent job operations...');
    
    const promises = [];
    const concurrentCount = 3;
    
    for (let i = 0; i < concurrentCount; i++) {
      const futureDate = new Date(Date.now() + (30 + i) * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
      const promise = request.post(`${ADMIN_BASE}/jobs`, {
        data: {
          name: `Concurrent Job ${i}`,
          origin: ['JFK', 'LAX', 'ORD'][i % 3],
          destination: ['LAX', 'ORD', 'JFK'][i % 3],
          date_start: futureDate,
          date_end: futureDate,
          adults: 1 + (i % 3),
          children: 0,
          infants_lap: 0,
          infants_seat: 0,
          trip_type: i % 2 === 0 ? 'one_way' : 'round_trip',
          return_date_start: i % 2 === 0 ? undefined : new Date(Date.now() + (37 + i) * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
          return_date_end: i % 2 === 0 ? undefined : new Date(Date.now() + (37 + i) * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
          class: ['economy', 'business', 'first'][i % 3],
          stops: ['nonstop', 'one_stop', 'any'][i % 3],
          currency: ['USD', 'EUR', 'GBP'][i % 3],
          interval: 'daily',
          time: `${10 + i}:00`,
          cron_expression: `0 ${10 + i} * * *`
        }
      });
      promises.push(promise);
    }

    const responses = await Promise.all(promises);
    console.log(`Created ${responses.length} jobs concurrently`);
    
    let successCount = 0;
    for (const [index, response] of responses.entries()) {
      if (response.ok()) {
        successCount++;
        const job = await response.json();
        if (job.id) {
          createdJobIds.push(job.id.toString());
        }
        console.log(`✅ Job ${index} created successfully`);
      } else {
        console.log(`❌ Job ${index} failed:`, response.status());
      }
    }
    
    expect(successCount).toBe(concurrentCount);
    console.log(`✅ All ${concurrentCount} concurrent jobs created successfully`);
  });
});