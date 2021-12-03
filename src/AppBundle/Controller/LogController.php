<?php

namespace AppBundle\Controller;

use AppBundle\Entity\Logs;
use AppBundle\Form\LogsType;
use AppBundle\Repository\LogRepository;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\Cache\Adapter\AdapterInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

/**
 * @Route("/")
 */
class LogController extends AbstractController
{
    /**
     * @var array $data
     */
    protected $data;

    /**
     * @var LogRepository $repo
     */
    private $repo;

    /**
     * @var EntityManagerInterface $entity
     */
    private $entity;

    /**
     * @var AdapterInterface $cache
     */
    private $cache;

    public function __construct(EntityManagerInterface $entity, LogRepository $repo, AdapterInterface $cache)
    {
        $this->entity = $entity;
        $this->repo = $repo;
        $this->cache = $cache;
    }

    /**
     * @Route("/", name="logs_index", methods={"GET"})
     * @throws \Doctrine\DBAL\Exception
     * @throws \Doctrine\DBAL\Driver\Exception
     */
    public function index(Request $request): Response
    {
//        $logs = $entityManager
//            ->getRepository(Logs::class)
//            ->findBy(
//                [],                 // $where
//                ['date' => 'ASC'],  // $orderBy
//                100,           // $limit
//                0             // $offset
//            );

        $search = trim($request->query->get('search'));
        $sql = $this->repo->getContent('Search.sql');

        $stmt = $this->entity
            ->getConnection()
            ->prepare($sql);
        $stmt->bindValue('search', "%$search%");

        $key = md5($sql . serialize($request->query->all()));

        $item = $this->cache->getItem($key);

        if ($item->isHit() && $this->getParameter('cache')) {
            $logs = $item->get();
        } else {
            $stmt->executeQuery();
            $logs = $stmt->fetchAll();
            $item->set($logs);
            $this->cache->save($item);
        }

        return $this->render('@App/logs/index.html.twig', [
            'search' => $search,
            'logs' => $logs,
        ]);
    }

    /**
     * @Route("/search", name="logs_search", methods={"GET"})
     * @throws \Doctrine\DBAL\Exception
     * @throws \Doctrine\DBAL\Driver\Exception
     */
    public function search(Request $request): Response
    {
        $url = trim($request->query->get('request'));
        $param = trim($request->query->get('param'));
        $date = trim($request->query->get('date'));

        $sql = $this->repo->getContent('Request.sql');

//        dump($this->validateDate($date));
//        if ($this->validateDate($date)) {
//            $sql = $this->repo->getContent('RequestDaily.sql');
//        }

        $stmt = $this->entity
            ->getConnection()
            ->prepare($sql);
        $stmt->bindValue('request', $url);
        $stmt->bindValue('param', $param);
        $stmt->bindValue('date', "$date%");

        $key = md5($sql . serialize($request->query->all()));

        $item = $this->cache->getItem($key);

        if ($item->isHit() && $this->getParameter('cache')) {
            $logs = $item->get();
        } else {
            $stmt->executeQuery();
            $logs = $stmt->fetchAll();
            $item->set($logs);
            $this->cache->save($item);
        }

//        dump($request->query->all());

        return $this->render('@App/search/index.html.twig', [
            'search' => $request->query->all(),
            'logs' => $logs,
        ]);
    }
    /**
     * @Route("/new", name="logs_new", methods={"GET", "POST"})
     */
    public function new(Request $request, EntityManagerInterface $entityManager): Response
    {
        $log = new Logs();
        $form = $this->createForm(LogsType::class, $log);
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {
            $entityManager->persist($log);
            $entityManager->flush();

            return $this->redirectToRoute('logs_index', [], Response::HTTP_SEE_OTHER);
        }

        return $this->render('logs/new.html.twig', [
            'log' => $log,
            'form' => $form->createView(),
        ]);
    }

    /**
     * @Route("/{id}", name="logs_show", methods={"GET"})
     */
    public function show(Logs $log): Response
    {
        return $this->render('logs/show.html.twig', [
            'log' => $log,
        ]);
    }

    /**
     * @Route("/{id}/edit", name="logs_edit", methods={"GET", "POST"})
     */
    public function edit(Request $request, Logs $log, EntityManagerInterface $entityManager): Response
    {
        $form = $this->createForm(LogsType::class, $log);
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {
            $entityManager->flush();

            return $this->redirectToRoute('logs_index', [], Response::HTTP_SEE_OTHER);
        }

        return $this->render('logs/edit.html.twig', [
            'log' => $log,
            'form' => $form->createView(),
        ]);
    }

    /**
     * @Route("/{id}", name="logs_delete", methods={"POST"})
     */
    public function delete(Request $request, Logs $log, EntityManagerInterface $entityManager): Response
    {
        if ($this->isCsrfTokenValid('delete'.$log->getId(), $request->request->get('_token'))) {
            $entityManager->remove($log);
            $entityManager->flush();
        }

        return $this->redirectToRoute('logs_index', [], Response::HTTP_SEE_OTHER);
    }

    protected function validateDate($date) {
        $format = 'Y-m-d';
        $d = \DateTime::createFromFormat($format, $date);

        return $d && $d->format($format) === $date;
    }
}
