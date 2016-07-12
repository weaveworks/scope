import collections
import json
import logging
import operator
import re

import flask
from oauth2client.client import GoogleCredentials
from googleapiclient import discovery

from google.appengine.api import urlfetch
from google.appengine.ext import ndb

app = flask.Flask('scheduler')
app.debug = True

# We use exponential moving average to record
# test run times.  Higher alpha discounts historic
# observations faster.
alpha = 0.3

class Test(ndb.Model):
  total_run_time = ndb.FloatProperty(default=0.) # Not total, but a EWMA
  total_runs = ndb.IntegerProperty(default=0)

  def parallelism(self):
    name = self.key.string_id()
    m = re.search('(\d+)_test.sh$', name)
    if m is None:
      return 1
    else:
      return int(m.group(1))

  def cost(self):
    p = self.parallelism()
    logging.info("Test %s has parallelism %d and avg run time %s", self.key.string_id(), p, self.total_run_time)
    return self.parallelism() * self.total_run_time

class Schedule(ndb.Model):
  shards = ndb.JsonProperty()

@app.route('/record/<path:test_name>/<runtime>', methods=['POST'])
@ndb.transactional
def record(test_name, runtime):
  test = Test.get_by_id(test_name)
  if test is None:
    test = Test(id=test_name)
  test.total_run_time = (test.total_run_time * (1-alpha)) + (float(runtime) * alpha)
  test.total_runs += 1
  test.put()
  return ('', 204)

@app.route('/schedule/<test_run>/<int:shard_count>/<int:shard>', methods=['POST'])
def schedule(test_run, shard_count, shard):
  # read tests from body
  test_names = flask.request.get_json(force=True)['tests']

  # first see if we have a scedule already
  schedule_id = "%s-%d" % (test_run, shard_count)
  schedule = Schedule.get_by_id(schedule_id)
  if schedule is not None:
    return flask.json.jsonify(tests=schedule.shards[str(shard)])

  # if not, do simple greedy algorithm
  test_times = ndb.get_multi(ndb.Key(Test, test_name) for test_name in test_names)
  def avg(test):
    if test is not None:
      return test.cost()
    return 1
  test_times = [(test_name, avg(test)) for test_name, test in zip(test_names, test_times)]
  test_times_dict = dict(test_times)
  test_times.sort(key=operator.itemgetter(1))

  shards = {i: [] for i in xrange(shard_count)}
  while test_times:
    test_name, time = test_times.pop()

    # find shortest shard and put it in that
    s, _ = min(((i, sum(test_times_dict[t] for t in shards[i]))
      for i in xrange(shard_count)), key=operator.itemgetter(1))

    shards[s].append(test_name)

  # atomically insert or retrieve existing schedule
  schedule = Schedule.get_or_insert(schedule_id, shards=shards)
  return flask.json.jsonify(tests=schedule.shards[str(shard)])

NAME_RE = re.compile(r'^host(?P<index>\d+)-(?P<build>\d+)-(?P<shard>\d+)$')

PROJECTS = [
  ('weaveworks/weave', 'positive-cocoa-90213', 'us-central1-a'),
  ('weaveworks/scope', 'scope-integration-tests', 'us-central1-a'),
]

@app.route('/tasks/gc')
def gc():
  # Get list of running VMs, pick build id out of VM name
  credentials = GoogleCredentials.get_application_default()
  compute = discovery.build('compute', 'v1', credentials=credentials)

  for repo, project, zone in PROJECTS:
    gc_project(compute, repo, project, zone)

  return "Done"

def gc_project(compute, repo, project, zone):
  logging.info("GCing %s, %s, %s", repo, project, zone)
  instances = compute.instances().list(project=project, zone=zone).execute()
  if 'items' not in instances:
    return

  host_by_build = collections.defaultdict(list)
  for instance in instances['items']:
    matches = NAME_RE.match(instance['name'])
    if matches is None:
      continue
    host_by_build[int(matches.group('build'))].append(instance['name'])
  logging.info("Running VMs by build: %r", host_by_build)

  # Get list of builds, filter down to runnning builds
  result = urlfetch.fetch('https://circleci.com/api/v1/project/%s' % repo,
    headers={'Accept': 'application/json'})
  assert result.status_code == 200
  builds = json.loads(result.content)
  running = {build['build_num'] for build in builds if not build.get('stop_time')}
  logging.info("Runnings builds: %r", running)

  # Stop VMs for builds that aren't running
  stopped = []
  for build, names in host_by_build.iteritems():
    if build in running:
      continue
    for name in names:
      stopped.append(name)
      logging.info("Stopping VM %s", name)
      compute.instances().delete(project=project, zone=zone, instance=name).execute()

  return
