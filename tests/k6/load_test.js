import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('critical_errors');
const isSmoke = __ENV.SMOKE === 'true';

export const options = {
  stages: isSmoke ? [
    { duration: '10s', target: 1 },
  ] : [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '2m', target: 50 },
    { duration: '30s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
  thresholds: isSmoke ? {
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    critical_errors: ['rate<0.05'],
  } : {
    http_req_duration: ['p(95)<500', 'p(99)<2000'],
    critical_errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

function generateID(prefix) {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

function smokeTestLogic() {
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'smoke: health status is 200': (r) => r.status === 200,
  });

  sleep(0.1);

  const teamName = generateID('smoke_team');
  const createTeamPayload = JSON.stringify({
    team_name: teamName,
    members: [
      { user_id: generateID('user'), username: 'User', is_active: true },
    ],
  });

  const createTeamRes = http.post(
    `${BASE_URL}/team/add`,
    createTeamPayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(createTeamRes, {
    'smoke: create team is 201': (r) => r.status === 201,
  });

  sleep(0.1);
}

export default function () {
  if (isSmoke) {
    smokeTestLogic();
    return;
  }

  const teamName = generateID('team');
  const authorID = generateID('user');
  const prID = generateID('pr');

  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
  });

  sleep(0.5);

  const createTeamPayload = JSON.stringify({
    team_name: teamName,
    members: [
      { user_id: `${authorID}`, username: 'Author', is_active: true },
      { user_id: `${authorID}_rev1`, username: 'Reviewer 1', is_active: true },
      { user_id: `${authorID}_rev2`, username: 'Reviewer 2', is_active: true },
      { user_id: `${authorID}_rev3`, username: 'Reviewer 3', is_active: true },
    ],
  });

  const createTeamRes = http.post(
    `${BASE_URL}/team/add`,
    createTeamPayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(createTeamRes, {
    'create team status is 201': (r) => r.status === 201,
  }) || errorRate.add(1);

  sleep(0.1);

  const getTeamRes = http.get(`${BASE_URL}/team/get?team_name=${teamName}`);
  check(getTeamRes, {
    'get team status is 200': (r) => r.status === 200,
  });

  sleep(0.1);

  const createPRPayload = JSON.stringify({
    pull_request_id: prID,
    pull_request_name: 'Load test PR',
    author_id: authorID,
  });

  const createPRRes = http.post(
    `${BASE_URL}/pullRequest/create`,
    createPRPayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  let reviewerID = null;

  check(createPRRes, {
    'create PR status is 201': (r) => r.status === 201,
  }) || errorRate.add(1);

  if (createPRRes.status === 201) {
    try {
      const body = JSON.parse(createPRRes.body);
      if (body.pr && body.pr.assigned_reviewers && body.pr.assigned_reviewers.length > 0) {
        reviewerID = body.pr.assigned_reviewers[0];
      }
    } catch (e) {
      // ignore parse error
    }
  }

  sleep(0.1);

  if (reviewerID) {
    const getReviewRes = http.get(`${BASE_URL}/users/getReview?user_id=${reviewerID}`);
    check(getReviewRes, {
      'get user reviews status is 200': (r) => r.status === 200,
    });
    sleep(0.1);
  }

  const mergePayload = JSON.stringify({
    pull_request_id: prID,
  });

  const mergeRes = http.post(
    `${BASE_URL}/pullRequest/merge`,
    mergePayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(mergeRes, {
    'merge PR status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(0.1);

  const mergeAgainRes = http.post(
    `${BASE_URL}/pullRequest/merge`,
    mergePayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(mergeAgainRes, {
    'merge PR again status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(0.2);
}

