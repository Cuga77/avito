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
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    critical_errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

function generateID(prefix) {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

export default function () {
  const teamName = generateID('team');
  const authorID = generateID('user');
  const prID = generateID('pr');

  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
    'health check has status ok': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'ok';
      } catch (e) {
        return false;
      }
    },
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
    'create team returns team_name': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.team && body.team.name === teamName;
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);

  sleep(0.5);


  const getTeamRes = http.get(`${BASE_URL}/team/get?team_name=${teamName}`);
  check(getTeamRes, {
    'get team status is 200': (r) => r.status === 200,
    'get team returns members': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.members && body.members.length === 4;
      } catch (e) {
        return false;
      }
    },
  });

  sleep(0.5);


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
    'create PR has reviewers': (r) => {
      try {
        const body = JSON.parse(r.body);
        if (body.pr && body.pr.assigned_reviewers && body.pr.assigned_reviewers.length > 0) {
          reviewerID = body.pr.assigned_reviewers[0];
          return true;
        }
        return false;
      } catch (e) {
        return false;
      }
    },
    'create PR author not in reviewers': (r) => {
      try {
        const body = JSON.parse(r.body);
        if (body.pr && body.pr.assigned_reviewers) {
          return !body.pr.assigned_reviewers.includes(authorID);
        }
        return true;
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);

  sleep(0.5);

  if (reviewerID) {
    const getReviewRes = http.get(`${BASE_URL}/users/getReview?user_id=${reviewerID}`);
    check(getReviewRes, {
      'get user reviews status is 200': (r) => r.status === 200,
      'get user reviews contains PR': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.pull_requests && body.pull_requests.length > 0;
        } catch (e) {
          return false;
        }
      },
    });

    sleep(0.5);

    const reassignPayload = JSON.stringify({
      pull_request_id: prID,
      old_user_id: reviewerID,
    });

    const reassignRes = http.post(
      `${BASE_URL}/pullRequest/reassign`,
      reassignPayload,
      { headers: { 'Content-Type': 'application/json' } }
    );

    check(reassignRes, {
      'reassign reviewer status is 200 or 400': (r) => r.status === 200 || r.status === 400,
      'reassign returns new reviewer if successful': (r) => {
        if (r.status === 200) {
          try {
            const body = JSON.parse(r.body);
            return body.new_reviewer_id !== reviewerID;
          } catch (e) {
            return false;
          }
        }
        return true;
      },
    });
  }

  sleep(0.5);

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
    'merge PR sets status to MERGED': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.pr && body.pr.status === 'MERGED';
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);

  sleep(0.5);

  const mergeAgainRes = http.post(
    `${BASE_URL}/pullRequest/merge`,
    mergePayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(mergeAgainRes, {
    'merge PR again status is 200': (r) => r.status === 200,
    'merge PR idempotent': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.pr && body.pr.status === 'MERGED';
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);

  sleep(0.5);

  if (reviewerID) {
    const reassignAfterMergePayload = JSON.stringify({
      pull_request_id: prID,
      old_user_id: reviewerID,
    });

    const reassignAfterMergeRes = http.post(
      `${BASE_URL}/pullRequest/reassign`,
      reassignAfterMergePayload,
      { headers: { 'Content-Type': 'application/json' } }
    );

    check(reassignAfterMergeRes, {
      'reassign after merge should fail': (r) => r.status !== 200,
    }) || errorRate.add(1);
  }

  sleep(2);
}

export function smokeTest() {
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'smoke: health check status is 200': (r) => r.status === 200,
  });
}

export function spikeTest() {
  const teamName = generateID('spike_team');
  const createTeamPayload = JSON.stringify({
    team_name: teamName,
    members: [
      { user_id: 'spike_user1', username: 'User 1', is_active: true },
      { user_id: 'spike_user2', username: 'User 2', is_active: true },
    ],
  });

  http.post(
    `${BASE_URL}/team/add`,
    createTeamPayload,
    { headers: { 'Content-Type': 'application/json' } }
  );
}

export function stressTest() {
  const operations = Math.floor(Math.random() * 3);

  switch (operations) {
    case 0:
      const teamName = generateID('stress_team');
      const createTeamPayload = JSON.stringify({
        team_name: teamName,
        members: [
          { user_id: generateID('user'), username: 'User', is_active: true },
        ],
      });
      http.post(`${BASE_URL}/team/add`, createTeamPayload, {
        headers: { 'Content-Type': 'application/json' },
      });
      break;

    case 1:
      http.get(`${BASE_URL}/health`);
      break;

    case 2:
      const prID = generateID('stress_pr');
      const authorID = generateID('stress_author');
      const createPRPayload = JSON.stringify({
        pull_request_id: prID,
        pull_request_name: 'Stress test PR',
        author_id: authorID,
      });
      http.post(`${BASE_URL}/pullRequest/create`, createPRPayload, {
        headers: { 'Content-Type': 'application/json' },
      });
      break;
  }
}
