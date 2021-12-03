select
    Date,
       Method,
       Request,
       Param,
       Port,
       Status,
       avg(Duration) as Duration,
       count(*)      as Number
from Logs
where Request = :request
  and Param = :param
and Date like :date
group by Date, Method, Request, Param, Port, Status
order by Date desc, Number desc;