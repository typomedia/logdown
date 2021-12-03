select
    Time as Date,
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
group by Time, Method, Request, Param, Port, Status
order by Time desc, Number desc;