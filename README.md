Marathon Consul Service Discovery
============

Publishes marathon services into consul.

Subscribes to the marathon event bus to receive task life cycle events. Periodically polls marathon tasks list.

Topology
========

On every mesos slave, run this daemon together with a consul daemon. This service will discover the mesos slave identifier of the current
host and register all marathon services (well, their port0) that run locally with the local consul agent.

Failure Modes
=============

This software dies
------------------

The consul agent will mark our services as critical after the TTL expires.

Slave node isolated from the network
------------------------------------

Consul will notice node being unavailable and stop offering services advertised by that node.

Marathon unavailable
-------------

Periodical polls will fail, no events will come in via the event bus. As a result, the TTL in consul will not be renewed services will
stop being advertised in consul.

This should be fixable by adding health checks in consul. See known issues

Mesos master unavailable
------------------------

This software does not directly depend on the mesos master. However, state served by marathon might become stale?


Mesos slave unavailable
-----------------------

Marathon might fire `TASK_LOST` events and we unregister the service? Startup of this service depends on the mesos slave being available
to discover our mesos slave id.

Stability
=========

This is currently proof of concept level. We're treating all Mesos/Marathon/Marathon-Consul like it could die anytime. Any services offered
via this mechanism are already available via normal VMs/Machines and mesos serves as a supplement only. Consul advertises a mixture of normal
daemons and marathon-launched daemons.

Known issues
============

 * This is only registering consul services with a TTL set, no health checks.
   Leaving that for later since the marathon health check code / events seem to be a little complex and up for refactoring
 * The callback is never unregistered with marathon when our process ends.
 * We get callbacks for all the tasks running everywhere on every instance of this service performance? scaling?
 * Service definitions, once registered in consul will not be unregistered / updated when this service restarts with different configuration
